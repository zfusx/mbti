export interface QuestionOption {
  text: string;
  value: string;
}

export interface Question {
  id: number;
  dimension: string;
  text: string;
  options: QuestionOption[];
}

export interface QuestionsResponse {
  count: number;
  questions: Question[];
}

export interface AnswerPayload {
  questionId: number;
  value: string;
}

export interface SessionResponse {
  sessionId: string;
}

export interface PairBreakdown {
  left: string;
  right: string;
  leftScore: number;
  rightScore: number;
  leftPct: number;
  rightPct: number;
  gap: number;
  winner: string;
}

export interface ScoreResult {
  type: string;
  tally: Record<string, number>;
  pairs: Record<string, PairBreakdown>;
  completion: number;
  confidence: {
    minPairGap: number;
    rule: string;
  };
}

export interface SubmitAnswersResponse {
  sessionId: string;
  result: ScoreResult;
}

export interface ResultDocument {
  type: string;
  nickname: string;
  image: string;
  ratio: string;
  description: string;
  keywords: string[];
  matches: string[];
  careers: string[];
}

export interface ResultResponse {
  data: ResultDocument;
}
