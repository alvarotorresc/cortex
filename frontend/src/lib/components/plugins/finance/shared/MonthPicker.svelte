<script lang="ts">
  import ChevronLeft from 'lucide-svelte/icons/chevron-left';
  import ChevronRight from 'lucide-svelte/icons/chevron-right';

  interface Props {
    month: string;
    onchange: (month: string) => void;
  }

  const { month, onchange }: Props = $props();

  const monthLabel = $derived(() => {
    const [year, m] = month.split('-').map(Number);
    const date = new Date(year, m - 1);
    return date.toLocaleDateString(undefined, { year: 'numeric', month: 'long' });
  });

  function navigate(delta: number): void {
    const [year, m] = month.split('-').map(Number);
    const date = new Date(year, m - 1 + delta);
    const nextMonth = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;
    onchange(nextMonth);
  }
</script>

<div class="flex items-center gap-2">
  <button
    onclick={() => navigate(-1)}
    class="rounded-[var(--radius-md)] p-2 text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
    aria-label="Previous month"
  >
    <ChevronLeft size={20} />
  </button>
  <span class="min-w-[160px] text-center text-sm font-medium text-[var(--color-text-primary)]">
    {monthLabel()}
  </span>
  <button
    onclick={() => navigate(1)}
    class="rounded-[var(--radius-md)] p-2 text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
    aria-label="Next month"
  >
    <ChevronRight size={20} />
  </button>
</div>
