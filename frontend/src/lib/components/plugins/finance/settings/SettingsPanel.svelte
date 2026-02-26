<script lang="ts">
  import X from 'lucide-svelte/icons/x';
  import { t } from 'svelte-i18n';

  interface Props {
    onclose: () => void;
  }

  const { onclose }: Props = $props();
</script>

<div class="space-y-4">
  <div class="flex items-center justify-between">
    <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">
      {$t('finance.settings')}
    </h3>
    <button
      onclick={onclose}
      class="rounded-[var(--radius-md)] p-2 text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
      aria-label={$t('finance.closeSettings')}
    >
      <X size={20} />
    </button>
  </div>

  <div
    class="flex flex-col items-center justify-center rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-6 py-16 text-center"
  >
    <p class="text-sm font-medium text-[var(--color-text-secondary)]">
      {$t('finance.settings')}
    </p>
    <p class="mt-1 text-xs text-[var(--color-text-tertiary)]">
      Categories, tags, accounts, recurring rules — coming soon
    </p>
  </div>
</div>
