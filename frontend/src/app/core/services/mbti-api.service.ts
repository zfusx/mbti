import { HttpClient, HttpParams } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import {
  AnswerPayload,
  QuestionsResponse,
  ResultResponse,
  SessionResponse,
  SubmitAnswersResponse,
} from '../models/mbti.models';

@Injectable({
  providedIn: 'root',
})
export class MbtiApiService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = environment.apiUrl;

  fetchQuestions(options?: { limit?: number; random?: boolean }): Observable<QuestionsResponse> {
    let params = new HttpParams();
    if (options?.limit) {
      params = params.set('limit', options.limit.toString());
    }
    if (options?.random) {
      params = params.set('random', String(options.random));
    }
    return this.http.get<QuestionsResponse>(`${this.baseUrl}/questions`, { params });
  }

  createSession(): Observable<SessionResponse> {
    return this.http.post<SessionResponse>(`${this.baseUrl}/sessions`, {});
  }

  submitAnswers(sessionId: string, answers: AnswerPayload[]): Observable<SubmitAnswersResponse> {
    return this.http.post<SubmitAnswersResponse>(`${this.baseUrl}/sessions/${sessionId}/answers`, {
      answers,
    });
  }

  fetchResult(type: string): Observable<ResultResponse> {
    return this.http.get<ResultResponse>(`${this.baseUrl}/results/${type}`);
  }
}
