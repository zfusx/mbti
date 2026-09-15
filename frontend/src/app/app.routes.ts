import { Routes } from '@angular/router';
import { Landing } from './features/landing/landing';
import { Quiz } from './features/quiz/quiz';
import { Results } from './features/results/results';

export const routes: Routes = [
  { path: '', component: Landing },
  { path: 'quiz', component: Quiz },
  { path: 'results/:type', component: Results },
  { path: '**', redirectTo: '', pathMatch: 'full' },
];
