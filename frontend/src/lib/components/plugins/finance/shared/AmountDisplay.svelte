<script lang="ts">
  interface Props {
    amount: number;
    currency?: string;
    showSign?: boolean;
  }

  const { amount, currency = 'EUR', showSign = false }: Props = $props();

  const formatted = $derived(() => {
    const abs = Math.abs(amount);
    const value = new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency,
    }).format(abs);

    if (!showSign) return value;
    return amount >= 0 ? `+${value}` : `-${value}`;
  });

  const colorClass = $derived(
    amount >= 0 ? 'text-[var(--color-success)]' : 'text-[var(--color-error)]',
  );
</script>

<span class={colorClass}>
  {formatted()}
</span>
