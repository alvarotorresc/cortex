<script lang="ts">
  interface Props {
    percentage: number;
    label?: string;
  }

  const { percentage, label }: Props = $props();

  const clampedWidth = $derived(Math.min(percentage, 100));

  const barColor = $derived(() => {
    if (percentage > 100) return 'bg-[var(--color-error)]';
    if (percentage >= 75) return 'bg-[var(--color-warning)]';
    return 'bg-[var(--color-success)]';
  });
</script>

<div class="space-y-1">
  {#if label}
    <div class="flex items-center justify-between">
      <span class="text-xs text-[var(--color-text-secondary)]">{label}</span>
      <span class="text-xs font-medium text-[var(--color-text-primary)]">
        {Math.round(percentage)}%
      </span>
    </div>
  {/if}
  <div
    class="h-2 w-full overflow-hidden rounded-[var(--radius-sm)] bg-[var(--color-bg-tertiary)]"
    role="progressbar"
    aria-valuenow={Math.round(percentage)}
    aria-valuemin={0}
    aria-valuemax={100}
    aria-label={label ?? `${Math.round(percentage)}% progress`}
  >
    <div
      class="h-full rounded-[var(--radius-sm)] transition-all duration-300 {barColor()}"
      style="width: {clampedWidth}%"
    ></div>
  </div>
</div>
