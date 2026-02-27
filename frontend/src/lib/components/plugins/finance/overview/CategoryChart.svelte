<script lang="ts">
  import { t } from 'svelte-i18n';
  import { Chart, registerables } from 'chart.js';
  import PieChart from 'lucide-svelte/icons/pie-chart';
  import type { CategoryTotal } from '../types';

  Chart.register(...registerables);

  interface Props {
    categories: CategoryTotal[];
  }

  const { categories }: Props = $props();

  let canvas = $state<HTMLCanvasElement | null>(null);
  let chartInstance: Chart | null = null;

  const CHART_COLORS = [
    '#0070F3', // brand-blue
    '#E11D48', // brand-red
    '#10B981', // emerald
    '#F59E0B', // amber
    '#8B5CF6', // violet
    '#EC4899', // pink
    '#06B6D4', // cyan
    '#F97316', // orange
    '#14B8A6', // teal
    '#6366F1', // indigo
    '#84CC16', // lime
    '#A855F7', // purple
  ];

  function getColors(count: number): string[] {
    const colors: string[] = [];
    for (let i = 0; i < count; i++) {
      colors.push(CHART_COLORS[i % CHART_COLORS.length]);
    }
    return colors;
  }

  $effect(() => {
    if (!canvas || categories.length === 0) return;

    chartInstance?.destroy();

    const colors = getColors(categories.length);

    chartInstance = new Chart(canvas, {
      type: 'doughnut',
      data: {
        labels: categories.map((c) => c.category || $t('finance.overview.uncategorized')),
        datasets: [
          {
            data: categories.map((c) => Math.abs(c.total)),
            backgroundColor: colors,
            borderWidth: 0,
            hoverBorderWidth: 2,
            hoverBorderColor: 'rgba(255,255,255,0.4)',
          },
        ],
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
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
                const value = context.parsed;
                const total = categories.reduce((sum, c) => sum + Math.abs(c.total), 0);
                const pct = total > 0 ? ((value / total) * 100).toFixed(1) : '0';
                const formatted = new Intl.NumberFormat(undefined, {
                  style: 'currency',
                  currency: 'EUR',
                }).format(value);
                return `${context.label}: ${formatted} (${pct}%)`;
              },
            },
          },
        },
        cutout: '65%',
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
    {$t('finance.overview.expensesByCategory')}
  </h3>

  {#if categories.length === 0}
    <div class="flex flex-col items-center justify-center py-8 text-center">
      <div class="mb-2 text-[var(--color-text-tertiary)]">
        <PieChart size={24} />
      </div>
      <p class="text-sm text-[var(--color-text-tertiary)]">
        {$t('finance.overview.noExpenses')}
      </p>
    </div>
  {:else}
    <div class="h-64">
      <canvas bind:this={canvas}></canvas>
    </div>
  {/if}
</div>
