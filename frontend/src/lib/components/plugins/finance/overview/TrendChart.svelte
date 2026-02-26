<script lang="ts">
  import { t } from 'svelte-i18n';
  import { Chart, registerables } from 'chart.js';
  import TrendingUp from 'lucide-svelte/icons/trending-up';
  import type { TrendPoint } from '../types';

  Chart.register(...registerables);

  interface Props {
    trends: TrendPoint[];
  }

  const { trends }: Props = $props();

  let canvas = $state<HTMLCanvasElement | null>(null);
  let chartInstance: Chart | null = null;

  function formatMonthLabel(month: string): string {
    const [year, m] = month.split('-');
    const date = new Date(Number(year), Number(m) - 1);
    return date.toLocaleDateString(undefined, { month: 'short', year: '2-digit' });
  }

  $effect(() => {
    if (!canvas || trends.length === 0) return;

    chartInstance?.destroy();

    const labels = trends.map((tp) => formatMonthLabel(tp.month));

    chartInstance = new Chart(canvas, {
      type: 'line',
      data: {
        labels,
        datasets: [
          {
            label: $t('finance.income'),
            data: trends.map((tp) => tp.income),
            borderColor: '#10B981',
            backgroundColor: 'rgba(16, 185, 129, 0.1)',
            fill: false,
            tension: 0.3,
            pointRadius: 4,
            pointHoverRadius: 6,
            borderWidth: 2,
          },
          {
            label: $t('finance.expense'),
            data: trends.map((tp) => tp.expense),
            borderColor: '#E11D48',
            backgroundColor: 'rgba(225, 29, 72, 0.1)',
            fill: false,
            tension: 0.3,
            pointRadius: 4,
            pointHoverRadius: 6,
            borderWidth: 2,
          },
          {
            label: $t('finance.balance'),
            data: trends.map((tp) => tp.balance),
            borderColor: '#0070F3',
            backgroundColor: 'rgba(0, 112, 243, 0.1)',
            fill: true,
            tension: 0.3,
            pointRadius: 4,
            pointHoverRadius: 6,
            borderWidth: 2,
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        interaction: {
          mode: 'index',
          intersect: false,
        },
        plugins: {
          legend: {
            position: 'bottom',
            labels: {
              color: 'rgb(156,163,175)',
              padding: 16,
              usePointStyle: true,
              pointStyleWidth: 10,
              font: { size: 12 },
            },
          },
          tooltip: {
            callbacks: {
              label(context) {
                const formatted = new Intl.NumberFormat(undefined, {
                  style: 'currency',
                  currency: 'EUR',
                }).format(context.parsed.y);
                return `${context.dataset.label}: ${formatted}`;
              },
            },
          },
        },
        scales: {
          x: {
            grid: {
              color: 'rgba(156,163,175,0.1)',
            },
            ticks: {
              color: 'rgb(156,163,175)',
              font: { size: 11 },
            },
          },
          y: {
            grid: {
              color: 'rgba(156,163,175,0.1)',
            },
            ticks: {
              color: 'rgb(156,163,175)',
              font: { size: 11 },
              callback(value) {
                return new Intl.NumberFormat(undefined, {
                  style: 'currency',
                  currency: 'EUR',
                  notation: 'compact',
                }).format(Number(value));
              },
            },
          },
        },
      },
    });

    return () => {
      chartInstance?.destroy();
      chartInstance = null;
    };
  });
</script>

<div
  class="rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5"
>
  <h3 class="mb-4 text-sm font-medium text-[var(--color-text-secondary)]">
    {$t('finance.overview.trends')}
  </h3>

  {#if trends.length === 0}
    <div class="flex flex-col items-center justify-center py-8 text-center">
      <div class="mb-2 text-[var(--color-text-tertiary)]">
        <TrendingUp size={24} />
      </div>
      <p class="text-sm text-[var(--color-text-tertiary)]">
        {$t('finance.overview.noTrends')}
      </p>
    </div>
  {:else}
    <div class="h-64">
      <canvas bind:this={canvas}></canvas>
    </div>
  {/if}
</div>
