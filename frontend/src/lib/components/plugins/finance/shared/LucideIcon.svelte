<script lang="ts">
  import Banknote from 'lucide-svelte/icons/banknote';
  import ShoppingCart from 'lucide-svelte/icons/shopping-cart';
  import Car from 'lucide-svelte/icons/car';
  import Gamepad2 from 'lucide-svelte/icons/gamepad-2';
  import Utensils from 'lucide-svelte/icons/utensils';
  import Receipt from 'lucide-svelte/icons/receipt';
  import HeartPulse from 'lucide-svelte/icons/heart-pulse';
  import CircleDot from 'lucide-svelte/icons/circle-dot';
  import Landmark from 'lucide-svelte/icons/landmark';
  import Wallet from 'lucide-svelte/icons/wallet';
  import TrendingUp from 'lucide-svelte/icons/trending-up';
  import PiggyBank from 'lucide-svelte/icons/piggy-bank';
  import Target from 'lucide-svelte/icons/target';
  import Home from 'lucide-svelte/icons/home';
  import GraduationCap from 'lucide-svelte/icons/graduation-cap';
  import Plane from 'lucide-svelte/icons/plane';
  import Gift from 'lucide-svelte/icons/gift';
  import Briefcase from 'lucide-svelte/icons/briefcase';
  import Wifi from 'lucide-svelte/icons/wifi';
  import Smartphone from 'lucide-svelte/icons/smartphone';
  import CreditCard from 'lucide-svelte/icons/credit-card';
  import CircleHelp from 'lucide-svelte/icons/circle-help';

  interface Props {
    name: string;
    size?: number;
  }

  const { name, size = 16 }: Props = $props();

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const iconMap: Record<string, any> = {
    banknote: Banknote,
    'shopping-cart': ShoppingCart,
    car: Car,
    'gamepad-2': Gamepad2,
    utensils: Utensils,
    receipt: Receipt,
    'heart-pulse': HeartPulse,
    'circle-dot': CircleDot,
    landmark: Landmark,
    wallet: Wallet,
    'trending-up': TrendingUp,
    'piggy-bank': PiggyBank,
    target: Target,
    home: Home,
    'graduation-cap': GraduationCap,
    plane: Plane,
    gift: Gift,
    briefcase: Briefcase,
    wifi: Wifi,
    smartphone: Smartphone,
    'credit-card': CreditCard,
  };

  const IconComponent = $derived(iconMap[name] ?? CircleHelp);
</script>

<IconComponent {size} />
