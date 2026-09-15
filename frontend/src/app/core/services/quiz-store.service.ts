import { Injectable, computed, effect, inject, signal } from '@angular/core';
import { firstValueFrom } from 'rxjs';
import { MbtiApiService } from './mbti-api.service';
import { AnswerPayload, Question, ScoreResult, SubmitAnswersResponse } from '../models/mbti.models';

type StoreStatus = 'idle' | 'loading' | 'ready' | 'error' | 'submitting' | 'completed';

@Injectable({
  providedIn: 'root',
})
export class QuizStore {
  private readonly api = inject(MbtiApiService);

  private initialized = false;
  private readonly questionsSignal = signal<Question[]>([]);
  private readonly statusSignal = signal<StoreStatus>('idle');
  private readonly errorSignal = signal<string | null>(null);
  private readonly currentIndexSignal = signal(0);
  private readonly answersSignal = signal<Map<number, string>>(new Map());
  private readonly sessionIdSignal = signal<string | null>(null);
  private readonly resultSignal = signal<ScoreResult | null>(null);
  private readonly questionLimitSignal = signal(92);

  readonly status = this.statusSignal.asReadonly();
  readonly error = this.errorSignal.asReadonly();
  readonly currentIndex = this.currentIndexSignal.asReadonly();
  readonly questions = this.questionsSignal.asReadonly();
  readonly answers = this.answersSignal.asReadonly();
  readonly result = this.resultSignal.asReadonly();
  readonly questionLimit = this.questionLimitSignal.asReadonly();

  readonly answeredCount = computed(() => this.answersSignal().size);
  readonly isComplete = computed(
    () =>
      this.questionsSignal().length > 0 &&
      this.answersSignal().size === this.questionsSignal().length,
  );
  readonly progress = computed(() => {
    const total = this.questionsSignal().length;
    if (!total) {
      return 0;
    }
    return Math.round((this.answeredCount() / total) * 100);
  });
  readonly currentQuestion = computed(
    () => this.questionsSignal()[this.currentIndexSignal()] ?? null,
  );

  constructor() {
    // If the questions array changes (after reload), reset transient state.
    effect(() => {
      const total = this.questionsSignal().length;
      if (total >= 0) {
        this.currentIndexSignal.set(0);
        this.answersSignal.set(new Map());
        this.resultSignal.set(null);
        this.sessionIdSignal.set(null);
      }
    });
  }

  async initialize(options?: { random?: boolean; limit?: number; force?: boolean }) {
    if (this.initialized && !options?.force) {
      if (this.statusSignal() === 'completed') {
        this.reset();
        return;
      }
      if (this.statusSignal() !== 'error') {
        return;
      }
    }
    this.initialized = true;
    if (options?.limit) {
      this.questionLimitSignal.set(options.limit);
    }
    await this.loadQuestions({ random: options?.random, limit: this.questionLimitSignal() });
  }

  private async loadQuestions(options?: { random?: boolean; limit?: number }) {
    this.statusSignal.set('loading');
    this.errorSignal.set(null);
    try {
      const response = await firstValueFrom(
        this.api.fetchQuestions({
          random: options?.random,
          limit: options?.limit ?? this.questionLimitSignal(),
        }),
      );
      this.questionsSignal.set(response.questions);
      this.statusSignal.set('ready');
    } catch (error) {
      console.error(error);
      this.errorSignal.set('Unable to load questions. Please try again.');
      this.statusSignal.set('error');
    }
  }

  selectAnswer(questionId: number, value: string) {
    const clone = new Map(this.answersSignal());
    clone.set(questionId, value);
    this.answersSignal.set(clone);
  }

  goToQuestion(index: number) {
    const questions = this.questionsSignal();
    if (!questions.length) {
      return;
    }
    const nextIndex = Math.max(0, Math.min(index, questions.length - 1));
    this.currentIndexSignal.set(nextIndex);
  }

  nextQuestion() {
    this.goToQuestion(this.currentIndexSignal() + 1);
  }

  previousQuestion() {
    this.goToQuestion(this.currentIndexSignal() - 1);
  }

  reset() {
    this.answersSignal.set(new Map());
    this.currentIndexSignal.set(0);
    this.statusSignal.set('ready');
    this.resultSignal.set(null);
    this.sessionIdSignal.set(null);
    this.errorSignal.set(null);
  }

  async setQuestionLimit(limit: number) {
    if (!limit || limit <= 0) {
      return;
    }
    if (this.questionLimitSignal() === limit && this.questionsSignal().length) {
      return;
    }
    this.questionLimitSignal.set(limit);
    await this.loadQuestions({ random: true, limit });
  }

  private async ensureSessionId(): Promise<string | null> {
    const existing = this.sessionIdSignal();
    if (existing) {
      return existing;
    }
    try {
      const { sessionId } = await firstValueFrom(this.api.createSession());
      this.sessionIdSignal.set(sessionId);
      return sessionId;
    } catch (error) {
      console.error(error);
      this.errorSignal.set('Could not initialize session.');
      return null;
    }
  }

  private buildPayload(): AnswerPayload[] {
    const answers: AnswerPayload[] = [];
    const questions = this.questionsSignal();
    const selections = this.answersSignal();
    for (const question of questions) {
      const value = selections.get(question.id);
      if (value) {
        answers.push({
          questionId: question.id,
          value,
        });
      }
    }
    return answers;
  }

  async submit(): Promise<SubmitAnswersResponse | null> {
    if (!this.questionsSignal().length) {
      return null;
    }
    if (!this.isComplete()) {
      this.errorSignal.set('Please answer every question before submitting.');
      return null;
    }

    const sessionId = await this.ensureSessionId();
    if (!sessionId) {
      return null;
    }

    const payload = this.buildPayload();
    this.statusSignal.set('submitting');
    this.errorSignal.set(null);

    try {
      const response = await firstValueFrom(this.api.submitAnswers(sessionId, payload));
      this.resultSignal.set(response.result);
      this.statusSignal.set('completed');
      return response;
    } catch (error) {
      console.error(error);
      this.errorSignal.set('Submission failed. Please try again.');
      this.statusSignal.set('ready');
      return null;
    }
  }
}
