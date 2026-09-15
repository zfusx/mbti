import { TestBed } from '@angular/core/testing';
import { of, throwError } from 'rxjs';
import { MbtiApiService } from './mbti-api.service';
import { QuizStore } from './quiz-store.service';
import { Question, ScoreResult } from '../models/mbti.models';

function questions(): Question[] {
  const dimensions = ['EI', 'SN', 'TF', 'JP'];
  return dimensions.flatMap((dimension, dimensionIndex) =>
    Array.from({ length: 5 }, (_, index) => ({
      id: dimensionIndex * 5 + index + 1,
      dimension,
      text: `Question ${dimensionIndex * 5 + index + 1}`,
      options: [
        { text: 'Left', value: dimension[0] },
        { text: 'Right', value: dimension[1] },
      ],
    })),
  );
}

describe('QuizStore', () => {
  const score: ScoreResult = {
    type: 'ESTJ',
    tally: {},
    pairs: {},
    completion: 100,
    confidence: { minPairGap: 1, rule: 'tie→prefer-left' },
  };

  let api: jasmine.SpyObj<MbtiApiService>;
  let store: QuizStore;

  beforeEach(() => {
    api = jasmine.createSpyObj<MbtiApiService>('MbtiApiService', [
      'fetchQuestions',
      'createSession',
      'submitAnswers',
    ]);
    api.fetchQuestions.and.returnValue(of({ count: 20, questions: questions() }));
    api.createSession.and.returnValue(of({ sessionId: 'session-id' }));
    api.submitAnswers.and.returnValue(of({ sessionId: 'session-id', result: score }));

    TestBed.configureTestingModule({
      providers: [QuizStore, { provide: MbtiApiService, useValue: api }],
    });
    store = TestBed.inject(QuizStore);
  });

  it('resets a completed quiz when the quiz route is opened again', async () => {
    await store.initialize({ limit: 20, random: true });
    for (const question of store.questions()) {
      store.selectAnswer(question.id, question.options[0].value);
    }

    await store.submit();
    expect(store.status()).toBe('completed');

    await store.initialize({ limit: 20, random: true });
    expect(store.status()).toBe('ready');
    expect(store.answeredCount()).toBe(0);
    expect(store.result()).toBeNull();
  });

  it('retries question loading after an error', async () => {
    spyOn(console, 'error');
    api.fetchQuestions.and.returnValues(
      throwError(() => new Error('network unavailable')),
      of({ count: 20, questions: questions() }),
    );

    await store.initialize({ limit: 20 });
    await store.initialize({ limit: 20 });

    expect(api.fetchQuestions).toHaveBeenCalledTimes(2);
  });
});
