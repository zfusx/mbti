import { NgClass } from '@angular/common';
import { ChangeDetectionStrategy, Component, OnInit, computed, inject } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { QuizStore } from '../../core/services/quiz-store.service';

@Component({
  selector: 'app-quiz',
  standalone: true,
  imports: [RouterLink, NgClass],
  templateUrl: './quiz.html',
  styleUrl: './quiz.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class Quiz implements OnInit {
  protected readonly store = inject(QuizStore);
  private readonly router = inject(Router);
  protected readonly quizModes = [
    { label: 'Quick / 20', value: 20 },
    { label: 'Balanced / 40', value: 40 },
    { label: 'Full / 92', value: 92 },
  ];

  protected readonly canSubmit = computed(
    () => this.store.isComplete() && this.store.status() === 'ready',
  );

  protected readonly isBusy = computed(() => {
    const status = this.store.status();
    return status === 'loading' || status === 'submitting';
  });

  protected readonly activeLimit = computed(() => this.store.questionLimit());

  async ngOnInit() {
    await this.store.initialize({ random: true, limit: this.activeLimit() });
  }

  protected selectOption(questionId: number, value: string) {
    if (this.isBusy()) {
      return;
    }
    this.store.selectAnswer(questionId, value);
  }

  protected selectedValue(questionId: number): string | undefined {
    return this.store.answers().get(questionId);
  }

  protected hasAnswer(questionId: number): boolean {
    return this.store.answers().has(questionId);
  }

  protected jumpTo(index: number) {
    this.store.goToQuestion(index);
  }

  protected next() {
    this.store.nextQuestion();
  }

  protected previous() {
    this.store.previousQuestion();
  }

  protected async submit() {
    const response = await this.store.submit();
    if (response) {
      this.router.navigate(['/results', response.result.type], {
        state: { result: response.result, sessionId: response.sessionId },
      });
    }
  }

  protected async pickMode(limit: number) {
    if (this.isBusy()) {
      return;
    }
    await this.store.setQuestionLimit(limit);
  }
}
