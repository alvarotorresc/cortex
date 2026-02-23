<script lang="ts">
  interface Props {
    size?: number;
    color?: string;
  }

  let { size = 32, color = 'currentColor' }: Props = $props();
</script>

<svg
  width={size}
  height={size}
  viewBox="0 0 32 32"
  fill="none"
  xmlns="http://www.w3.org/2000/svg"
  role="img"
  aria-label="Cortex logo"
>
  <!-- Central hexagon -->
  <path
    d="M16 6L22.9 10V18L16 22L9.1 18V10L16 6Z"
    stroke={color}
    stroke-width="1.5"
    stroke-linejoin="round"
  />
  <!-- Central node -->
  <circle cx="16" cy="14" r="2.5" fill={color} />
  <!-- Radiating connections with end nodes -->
  <!-- Top -->
  <line x1="16" y1="6" x2="16" y2="2" stroke={color} stroke-width="1.5" stroke-linecap="round" />
  <circle cx="16" cy="1.5" r="1.5" fill={color} />
  <!-- Top-right -->
  <line
    x1="22.9"
    y1="10"
    x2="26.3"
    y2="8"
    stroke={color}
    stroke-width="1.5"
    stroke-linecap="round"
  />
  <circle cx="26.8" cy="7.7" r="1.5" fill={color} />
  <!-- Bottom-right -->
  <line
    x1="22.9"
    y1="18"
    x2="26.3"
    y2="20"
    stroke={color}
    stroke-width="1.5"
    stroke-linecap="round"
  />
  <circle cx="26.8" cy="20.3" r="1.5" fill={color} />
  <!-- Bottom -->
  <line
    x1="16"
    y1="22"
    x2="16"
    y2="26"
    stroke={color}
    stroke-width="1.5"
    stroke-linecap="round"
  />
  <circle cx="16" cy="26.5" r="1.5" fill={color} />
  <!-- Bottom-left -->
  <line
    x1="9.1"
    y1="18"
    x2="5.7"
    y2="20"
    stroke={color}
    stroke-width="1.5"
    stroke-linecap="round"
  />
  <circle cx="5.2" cy="20.3" r="1.5" fill={color} />
  <!-- Top-left -->
  <line
    x1="9.1"
    y1="10"
    x2="5.7"
    y2="8"
    stroke={color}
    stroke-width="1.5"
    stroke-linecap="round"
  />
  <circle cx="5.2" cy="7.7" r="1.5" fill={color} />
</svg>
