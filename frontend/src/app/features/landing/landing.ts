import { ChangeDetectionStrategy, Component } from '@angular/core';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-landing',
  standalone: true,
  imports: [RouterLink],
  templateUrl: './landing.html',
  styleUrl: './landing.css',
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class Landing {
  protected readonly stats = [
    { label: 'Question Bank', value: '92 researched prompts spanning EI / SN / TF / JP' },
    { label: 'Processing', value: '< 2s scoring via Go 1.25 Chi service + JSON payloads' },
    { label: 'Storage', value: 'PostgreSQL 17 JSONB with GIN indexes ready for analytics' },
  ];

  protected readonly blueprint = [
    {
      title: 'Experience Layer',
      details: 'Angular 20 + signals, Vite dev server, and Tailwind tokens for sleek UI states.',
    },
    {
      title: 'Service Layer',
      details: 'Go API exposes /questions, /sessions, /results with deterministic scoring logic.',
    },
    {
      title: 'Insights Layer',
      details: 'PostgreSQL JSONB keeps raw answers + catalog data ready for personalization.',
    },
  ];
}
