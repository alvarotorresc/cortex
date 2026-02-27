<script lang="ts">
  import { t } from 'svelte-i18n';
  import type { InvestmentWithPnL, InvestmentType } from '../types';
  import AmountDisplay from '../shared/AmountDisplay.svelte';

  interface Props {
    investment: InvestmentWithPnL;
    onedit: (inv: InvestmentWithPnL) => void;
  }

  const { investment, onedit }: Props = $props();

  const typeBadgeColors: Record<InvestmentType, { bg: string; text: string }> = {
    crypto: { bg: 'rgba(249, 115, 22, 0.15)', text: 'rgb(249, 115, 22)' },
    stock: { bg: 'rgba(59, 130, 246, 0.15)', text: 'rgb(59, 130, 246)' },
    etf: { bg: 'rgba(168, 85, 247, 0.15)', text: 'rgb(168, 85, 247)' },
    fund: { bg: 'rgba(20, 184, 166, 0.15)', text: 'rgb(20, 184, 166)' },
    other: { bg: 'rgba(107, 114, 128, 0.15)', text: 'rgb(107, 114, 128)' },
  };

  const badgeColor = $derived(typeBadgeColors[investment.type]);

  const formattedLastUpdated = $derived(() => {
    if (!investment.last_updated) return null;
    const date = new Date(investment.last_updated);
    return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
  });

  const pnlPositive = $derived((investment.pnl ?? 0) >= 0);
</script>

<button
  type="button"
  onclick={() => onedit(investment)}
  class="flex w-full flex-col gap-3 rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-4 text-left transition-colors hover:bg-[var(--color-bg-tertiary)]"
>
  <!-- Header: Name + Type badge -->
  <div class="flex items-start justify-between gap-2">
    <h3 class="text-sm font-semibold text-[var(--color-text-primary)]">
      {investment.name}
    </h3>
    <span
      class="shrink-0 rounded-full px-2.5 py-0.5 text-[11px] font-semibold uppercase tracking-wide"
      style:background-color={badgeColor.bg}
      style:color={badgeColor.text}
    >
      {$t(`finance.investments.${investment.type}`)}
    </span>
  </div>

  <!-- Units & Avg Price -->
  {#if investment.units != null || investment.avg_buy_price != null}
    <div class="flex flex-wrap gap-x-4 gap-y-1 text-xs text-[var(--color-text-tertiary)]">
      {#if investment.units != null}
        <span>
          <span class="font-medium text-[var(--color-text-secondary)]">{$t('finance.investments.units')}:</span>
          {investment.units.toLocaleString(undefined, { maximumFractionDigits: 6 })}
        </span>
      {/if}
      {#if investment.avg_buy_price != null}
        <span>
          <span class="font-medium text-[var(--color-text-secondary)]">{$t('finance.investments.avgPrice')}:</span>
          {new Intl.NumberFormat(undefined, { style: 'currency', currency: investment.currency }).format(investment.avg_buy_price)}
        </span>
      {/if}
    </div>
  {/if}

  <!-- Current Price -->
  {#if investment.current_price != null}
    <div class="text-xs text-[var(--color-text-tertiary)]">
      <span class="font-medium text-[var(--color-text-secondary)]">{$t('finance.investments.currentPrice')}:</span>
      {new Intl.NumberFormat(undefined, { style: 'currency', currency: investment.currency }).format(investment.current_price)}
    </div>
  {/if}

  <!-- P&L row -->
  {#if investment.pnl != null}
    <div class="flex items-center justify-between border-t border-[var(--color-border)] pt-3">
      <span class="text-xs font-medium text-[var(--color-text-secondary)]">
        {$t('finance.investments.pnl')}
      </span>
      <div class="flex items-center gap-2">
        <span class="text-sm font-semibold">
          <AmountDisplay amount={investment.pnl} currency={investment.currency} showSign />
        </span>
        {#if investment.pnl_percentage != null}
          <span
            class="rounded-full px-2 py-0.5 text-[11px] font-semibold {pnlPositive
              ? 'text-[var(--color-success)]'
              : 'text-[var(--color-error)]'}"
            style:background-color={pnlPositive
              ? 'color-mix(in srgb, var(--color-success) 10%, transparent)'
              : 'color-mix(in srgb, var(--color-error) 10%, transparent)'}
          >
            {pnlPositive ? '+' : ''}{investment.pnl_percentage.toFixed(2)}%
          </span>
        {/if}
      </div>
    </div>
  {/if}

  <!-- Last updated -->
  {#if formattedLastUpdated()}
    <p class="text-[11px] text-[var(--color-text-tertiary)]">
      {$t('finance.investments.lastUpdated')}: {formattedLastUpdated()}
    </p>
  {/if}
</button>
