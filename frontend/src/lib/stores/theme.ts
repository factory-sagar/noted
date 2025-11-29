import { writable } from 'svelte/store';

export type Theme = 'modern' | 'minimal' | 'cyber' | 'noir' | 'retro' | 'nordic' | 'monokai' | 'dracula' | 'solarized' | 'ocean' | 'forest';

const allThemes = [
  'theme-modern', 
  'theme-minimal', 
  'theme-cyber', 
  'theme-noir',
  'theme-retro',
  'theme-nordic',
  'theme-monokai',
  'theme-dracula',
  'theme-solarized',
  'theme-ocean',
  'theme-forest'
];

const createThemeStore = () => {
  const { subscribe, set, update } = writable<Theme>('modern');

  return {
    subscribe,
    set: (theme: Theme) => {
      if (typeof window !== 'undefined') {
        localStorage.setItem('theme', theme);
        // Remove all theme classes
        document.documentElement.classList.remove(...allThemes);
        // Add the new one (if not modern)
        if (theme !== 'modern') {
          document.documentElement.classList.add(`theme-${theme}`);
        }
      }
      set(theme);
    },
    init: () => {
      if (typeof window !== 'undefined') {
        const saved = localStorage.getItem('theme') as Theme;
        if (saved && saved !== 'modern') {
          document.documentElement.classList.add(`theme-${saved}`);
          set(saved);
        }
      }
    }
  };
};

export const theme = createThemeStore();
