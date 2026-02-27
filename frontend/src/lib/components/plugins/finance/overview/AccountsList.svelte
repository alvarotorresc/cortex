<script lang="ts">
  import { t } from 'svelte-i18n';
  import Landmark from 'lucide-svelte/icons/landmark';
  import Banknote from 'lucide-svelte/icons/banknote';
  import Wallet from 'lucide-svelte/icons/wallet';
  import TrendingUp from 'lucide-svelte/icons/trending-up';
  import type { AccountWithBalance, AccountType } from '../types';
  import AmountDisplay from '../shared/AmountDisplay.svelte';

  interface Props {
    accounts: AccountWithBalance[];
  }

  const { accounts }: Props = $props();

  const activeAccounts = $derived(accounts.filter((a) => !a.is_archived));

  const TYPE_LABELS: Record<AccountType, string> = {
    checking: 'finance.overview.accountTypes.checking',
    savings: 'finance.overview.accountTypes.savings',
    cash: 'finance.overview.accountTypes.cash',
    investment: 'finance.overview.accountTypes.investment',
  };

  const TYPE_COLORS: Record<AccountType, string> = {
    checking: 'bg-[var(--color-brand-blue)]/15 text-[var(--color-brand-blue)]',
    savings: 'bg-[var(--color-success)]/15 text-[var(--color-success)]',
    cash: 'bg-[var(--color-warning)]/15 text-[var(--color-warning)]',
    investment: 'bg-purple-500/15 text-purple-400',
  };

  function getAccountIcon(type: AccountType) {
    switch (type) {
      case 'checking':
        return Landmark;
      case 'savings':
        return Banknote;
      case 'cash':
        return Wallet;
      case 'investment':
        return TrendingUp;
    }
  }
</script>

<div
  class="rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5"
>
  <h3 class="mb-4 text-sm font-medium text-[var(--color-text-secondary)]">
    {$t('finance.overview.accounts')}
  </h3>

  {#if activeAccounts.length === 0}
    <div class="flex flex-col items-center justify-center py-6 text-center">
      <p class="text-sm text-[var(--color-text-tertiary)]">
        {$t('finance.overview.noAccounts')}
      </p>
    </div>
  {:else}
    <div class="flex flex-col gap-2">
      {#each activeAccounts as account (account.id)}
        {@const IconComponent = getAccountIcon(account.type)}
        <div
          class="flex items-center justify-between rounded-[var(--radius-md)] bg-[var(--color-bg-tertiary)] px-4 py-3"
        >
          <div class="flex items-center gap-3">
            <div
              class="flex h-8 w-8 shrink-0 items-center justify-center rounded-[var(--radius-sm)] {TYPE_COLORS[account.type]}"
            >
              <IconComponent size={16} />
            </div>
            <div>
              <p class="text-sm font-medium text-[var(--color-text-primary)]">
                {account.name}
              </p>
              <p class="text-xs text-[var(--color-text-tertiary)]">
                {$t(TYPE_LABELS[account.type])}
              </p>
            </div>
          </div>
          <div class="text-sm font-semibold">
            <AmountDisplay amount={account.balance} currency={account.currency} />
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>
