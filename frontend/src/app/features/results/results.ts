import { ChangeDetectionStrategy, Component, computed, inject, signal } from '@angular/core';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import { catchError, filter, map, of, switchMap, tap } from 'rxjs';
import { MbtiApiService } from '../../core/services/mbti-api.service';
import { ResultDocument, ScoreResult } from '../../core/models/mbti.models';

@Component({
  selector: 'app-results',
  standalone: true,
  imports: [RouterLink],
  templateUrl: './results.html',
  styleUrl: './results.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class Results {
  private readonly api = inject(MbtiApiService);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  protected readonly loading = signal(true);
  protected readonly error = signal<string | null>(null);
  protected readonly document = signal<ResultDocument | null>(null);
  protected readonly type = signal('');
  protected readonly score = signal<ScoreResult | null>(
    (this.router.getCurrentNavigation()?.extras.state?.['result'] as ScoreResult) ??
      (history.state?.['result'] as ScoreResult) ??
      null,
  );

  protected readonly pairEntries = computed(() => {
    const pairs = this.score()?.pairs;
    return pairs ? Object.entries(pairs) : [];
  });

  protected readonly tallyEntries = computed(() => {
    const tally = this.score()?.tally;
    return tally ? Object.entries(tally) : [];
  });

  constructor() {
    this.route.paramMap
      .pipe(
        takeUntilDestroyed(),
        map((params) => (params.get('type') ?? '').toUpperCase()),
        tap((type) => {
          if (!type) {
            this.error.set('Result type missing.');
          } else {
            this.error.set(null);
            this.type.set(type);
            this.loading.set(true);
          }
        }),
        filter((type) => !!type),
        switchMap((type) =>
          this.api.fetchResult(type).pipe(
            catchError((err) => {
              console.error(err);
              this.error.set('Unable to load narrative copy.');
              this.loading.set(false);
              return of(null);
            }),
          ),
        ),
      )
      .subscribe((response) => {
        this.loading.set(false);
        if (response?.data) {
          this.document.set(response.data);
        }
      });
  }

  protected hasScore(): boolean {
    return !!this.score();
  }
}
