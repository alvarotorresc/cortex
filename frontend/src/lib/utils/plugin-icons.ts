import Wallet from 'lucide-svelte/icons/wallet';
import StickyNote from 'lucide-svelte/icons/sticky-note';
import Puzzle from 'lucide-svelte/icons/puzzle';
import NotebookPen from 'lucide-svelte/icons/notebook-pen';
import FolderGit2 from 'lucide-svelte/icons/folder-git-2';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const iconMap: Record<string, any> = {
  wallet: Wallet,
  'sticky-note': StickyNote,
  'notebook-pen': NotebookPen,
  'folder-git-2': FolderGit2,
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function getPluginIcon(iconName: string): any {
  return iconMap[iconName] ?? Puzzle;
}
