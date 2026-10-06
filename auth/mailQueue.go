package auth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type MailSender interface {
	SendVerificationMail(string, string) error
}

type MailJob struct {
	Token, To string
}

type MailQueueConfig struct {
	Workers, Retries, QueueSize int
	RetryDelay, MaxDelay        time.Duration
}

// не надо ли называть этот как провайде, адаптер? Более дженерик
type MailQueue struct {
	sender MailSender
	cfg    MailQueueConfig
	jobs   chan MailJob
	wg     sync.WaitGroup
	// Не факто что накладные расходы дают выигрыш при малых колиествах читателей
	mu     sync.RWMutex
	ctx    context.Context
	closed bool
	cancel context.CancelFunc
}

func NewMailQueue(cfg MailQueueConfig, sender MailSender) *MailQueue {
	// Когда вызывается Close канал закрывается, туда ничего не кладётся, но воркеры выполняют джобы полноценно
	// Когда контекст закрывается, воркеры в ожидании возвращаются через case, разбирают всё быстро через return.
	// Но сейчас у нас при вызове close передаётся контекст приложения который закрывается по сигналу от os,
	// То есть он уже закрыт, а наш контекст не должен на эту отмену реагировать
	// Чтобы сделать контекст не реагирующий на отмену используем WithoutCancel context.WithCancel(context.WithoutCancel(parent))
	// Просто свой собственный инициализировали
	newCtx, cancel := context.WithCancel(context.Background())

	q := &MailQueue{
		cfg:    cfg,
		jobs:   make(chan MailJob, cfg.QueueSize),
		ctx:    newCtx,
		sender: sender,
		cancel: cancel,
		// Понять какими значениями по умолчании инициализируются значения просто ZeroValue?
		// Типо WaitGroup, closed и так далее
	}

	for i := 0; i < cfg.Workers; i++ {
		q.wg.Add(1)
		go q.worker(i + 1)
	}

	return q
}

func (q *MailQueue) Close(ctx context.Context) error {
	q.mu.Lock()

	if !q.closed {
		q.closed = true
		close(q.jobs)
	}

	q.mu.Unlock()

	finishChan := make(chan int)

	go func() {
		q.wg.Wait()
		close(finishChan)
	}()

	select {
	case <-finishChan:
		// Тут чисто для конситстентности и перестраховки также закрою
		// хотя по сути все воркеры завершены
		q.cancel()
		return nil
	case <-ctx.Done():
		q.cancel()
		// тут контекст закрывается, чтобы прервать
		// все ожидания таймаутов на ретраи, если они есть,
		// и чтобы быстро проскипать все оставышиеся сообщения очереди через ctx.Err() return.
		return fmt.Errorf("MailQueue: Close: %w", ctx.Err())
	}
}

func (q *MailQueue) AddVerificationJob(to, token string) error {
	return q.enqueue(MailJob{Token: token, To: to})
}

func (q *MailQueue) enqueue(job MailJob) error {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.closed {
		return fmt.Errorf("MailQueue: enqueue: to: %s, %w", job.To, ErrMailQueueClosed)
	}

	select {
	case q.jobs <- job:
		return nil
	default:
		fmt.Println("queue is busy")
		return fmt.Errorf("MailQueue: enqueue: to: %s, %w", job.To, ErrMailQueueFull)
	}
}

func (q *MailQueue) worker(id int) {
	defer q.wg.Done()

	for job := range q.jobs {
		fmt.Printf("Wirker %d took job with to: %s", id, job.To)
		if err := q.recoverWrap(job); err != nil {
			// TODO тут ошибку некуда возвращать
			formatedErr := fmt.Errorf("MailQueue: worker: to: %s: %w", job.To, err)
			fmt.Println(formatedErr)
		}

	}
}

func (q *MailQueue) recoverWrap(job MailJob) (err error) {
	// В каждой горутине свой отлов паники
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf("MailQueue: recoverWrap: %v", panicErr)
		}
	}()

	return q.handleJob(job)
}

func (q *MailQueue) handleJob(job MailJob) error {
	// Если контекст приложения закрывается, то просто возвращаем ошибки не отправляем больше письма
	// Они теряются эт оне очень
	if q.ctx.Err() != nil {
		return fmt.Errorf("MailQueue: handleJob: to: %s: %w", job.To, q.ctx.Err())
	}

	for attempt := 1; attempt <= q.cfg.Retries; attempt++ {
		err := q.sender.SendVerificationMail(job.To, job.Token)

		if err == nil {
			return nil
		}

		// TODO jitter: разброс в пределах половины задержки
		if attempt < q.cfg.Retries {
			delay := min(q.cfg.MaxDelay, q.cfg.RetryDelay<<attempt)

			select {
			case <-time.After(delay):
				// Висим на ожидании таймаута
				continue
			case <-q.ctx.Done():
				// Если контекст прерывается сразу пролетаем сюда
				return fmt.Errorf("MailQueue: handleJob: to: %s: %w", job.To, err)
			}
		}

		return fmt.Errorf(
			"MailQueue: handleJob: to: %s: after %d retries: %w",
			job.To,
			q.cfg.Retries,
			err,
		)
	}

	return nil

}

// func (q *MailQueue)sendMail(job MailJob) error {

// }
