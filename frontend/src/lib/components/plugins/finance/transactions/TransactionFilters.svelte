<script lang="ts">
  import { t } from 'svelte-i18n';
  import Search from 'lucide-svelte/icons/search';
  import type { Category, AccountWithBalance, Tag, TransactionFilter } from '../types';

  interface Props {
    categories: Category[];
    accounts: AccountWithBalance[];
    tags: Tag[];
    onfilter: (filters: TransactionFilter) => void;
  }

  const { categories, accounts, tags, onfilter }: Props = $props();

  let accountFilter = $state('');
  let categoryFilter = $state('');
  let typeFilter = $state('');
  let tagFilter = $state('');
  let searchText = $state('');
  let debounceTimer = $state<ReturnType<typeof setTimeout> | undefined>(undefined);

  function emitFilters(): void {
    const filters: TransactionFilter = {};
    if (accountFilter) filters.account = accountFilter;
    if (categoryFilter) filters.category = categoryFilter;
    if (typeFilter) filters.type = typeFilter;
    if (tagFilter) filters.tag = tagFilter;
    if (searchText.trim()) filters.search = searchText.trim();
    onfilter(filters);
  }

  function handleSelectChange(): void {
    emitFilters();
  }

  function handleSearchInput(event: Event): void {
    const target = event.target as HTMLInputElement;
    searchText = target.value;
    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      emitFilters();
    }, 300);
  }
</script>

<div class="flex flex-wrap items-center gap-3">
  <!-- Search -->
  <div class="relative min-w-[200px] flex-1">
    <Search
      size={16}
      class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[var(--color-text-tertiary)]"
    />
    <input
      type="text"
      value={searchText}
      oninput={handleSearchInput}
      placeholder={$t('finance.search')}
      class="w-full rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] py-2 pl-9 pr-3 text-sm text-[var(--color-text-primary)] placeholder-[var(--color-text-tertiary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
    />
  </div>

  <!-- Account filter -->
  <select
    bind:value={accountFilter}
    onchange={handleSelectChange}
    aria-label={$t('finance.account')}
    class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
  >
    <option value="">{$t('finance.allAccounts')}</option>
    {#each accounts as account (account.id)}
      <option value={String(account.id)}>{account.name}</option>
    {/each}
  </select>

  <!-- Category filter -->
  <select
    bind:value={categoryFilter}
    onchange={handleSelectChange}
    aria-label={$t('finance.category')}
    class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
  >
    <option value="">{$t('finance.allCategories')}</option>
    {#each categories as category (category.id)}
      <option value={category.name}>{category.name}</option>
    {/each}
  </select>

  <!-- Type filter -->
  <select
    bind:value={typeFilter}
    onchange={handleSelectChange}
    aria-label={$t('finance.type')}
    class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
  >
    <option value="">{$t('finance.allTypes')}</option>
    <option value="income">{$t('finance.income')}</option>
    <option value="expense">{$t('finance.expense')}</option>
    <option value="transfer">{$t('finance.transfer')}</option>
  </select>

  <!-- Tag filter -->
  <select
    bind:value={tagFilter}
    onchange={handleSelectChange}
    aria-label={$t('finance.tags')}
    class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
  >
    <option value="">{$t('finance.allTags')}</option>
    {#each tags as tag (tag.id)}
      <option value={tag.name}>{tag.name}</option>
    {/each}
  </select>
</div>
