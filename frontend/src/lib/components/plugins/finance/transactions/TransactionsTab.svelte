<script lang="ts">
  import ArrowLeftRight from 'lucide-svelte/icons/arrow-left-right';

  interface Props {
    month: string;
  }

  const { month }: Props = $props();
</script>

<div
  class="flex flex-col items-center justify-center rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-6 py-16 text-center"
>
  <div class="mb-3 text-[var(--color-text-tertiary)]">
    <ArrowLeftRight size={32} />
  </div>
  <p class="text-sm font-medium text-[var(--color-text-secondary)]">Transactions</p>
  <p class="mt-1 text-xs text-[var(--color-text-tertiary)]">Coming soon ({month})</p>
</div>
