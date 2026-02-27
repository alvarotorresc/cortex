<script lang="ts">
  import type { Transaction, Category, AccountWithBalance } from '../types';
  import AmountDisplay from '../shared/AmountDisplay.svelte';

  interface Props {
    transaction: Transaction;
    categories: Category[];
    accounts: AccountWithBalance[];
    onedit: (tx: Transaction) => void;
  }

  const { transaction, categories, accounts, onedit }: Props = $props();

  const category = $derived(
    categories.find((c) => c.name === transaction.category),
  );

  const account = $derived(
    accounts.find((a) => a.id === transaction.account_id),
  );

  const displayAmount = $derived(
    transaction.type === 'expense' ? -transaction.amount : transaction.amount,
  );

  const formattedDate = $derived(() => {
    const date = new Date(transaction.date);
    return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
  });
</script>

<button
  type="button"
  onclick={() => onedit(transaction)}
  class="flex w-full items-center gap-3 rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-4 py-3 text-left transition-colors hover:bg-[var(--color-bg-tertiary)]"
>
  <!-- Date -->
  <span class="w-16 shrink-0 text-xs text-[var(--color-text-tertiary)]">
    {formattedDate()}
  </span>

  <!-- Category dot + Description -->
  <div class="flex min-w-0 flex-1 items-center gap-2">
    {#if category}
      <span
        class="inline-block h-2.5 w-2.5 shrink-0 rounded-full"
        style:background-color={category.color}
        aria-hidden="true"
      ></span>
    {/if}
    <div class="min-w-0 flex-1">
      <p class="truncate text-sm font-medium text-[var(--color-text-primary)]">
        {transaction.description}
      </p>
      <p class="truncate text-xs text-[var(--color-text-tertiary)]">
        {category?.name ?? transaction.category}
        {#if account}
          &middot; {account.name}
        {/if}
      </p>
    </div>
  </div>

  <!-- Tags -->
  {#if transaction.tags.length > 0}
    <div class="hidden shrink-0 items-center gap-1 sm:flex">
      {#each transaction.tags as tag (tag.id)}
        <span
          class="rounded-full px-2 py-0.5 text-[10px] font-medium"
          style:background-color="{tag.color}20"
          style:color={tag.color}
        >
          {tag.name}
        </span>
      {/each}
    </div>
  {/if}

  <!-- Amount -->
  <span class="shrink-0 text-sm font-semibold">
    <AmountDisplay amount={displayAmount} showSign />
  </span>
</button>
