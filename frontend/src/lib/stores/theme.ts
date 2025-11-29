import { writable } from 'svelte/store';

export type Theme = 'modern' | 'minimal' | 'cyber' | 'noir';

const createThemeStore = () => {
  const { subscribe, set, update } = writable<Theme>('modern');

  return {
    subscribe,
    set: (theme: Theme) => {
      if (typeof window !== 'undefined') {
        localStorage.setItem('theme', theme);
        // Remove all theme classes
        document.documentElement.classList.remove('theme-modern', 'theme-minimal', 'theme-cyber', 'theme-noir');
        // Add the new one (unless it's modern/default which has no class in my CSS design, but let's be explicit if needed or just rely on root)
        // My CSS design uses :root for modern, so we don't STRICTLY need a class for it, but it's good practice to clean up others.
        // If theme is not modern, add the class.
        if (theme !== 'modern') {
          document.documentElement.classList.add(`theme-${theme}`);
        }
      }
      set(theme);
    },
    init: () => {
      if (typeof window !== 'undefined') {
        const saved = localStorage.getItem('theme') as Theme;
        if (saved) {
          if (saved !== 'modern') {
            document.documentElement.classList.add(`theme-${saved}`);
          }
          set(saved);
        }
      }
    }
  };
};

export const theme = createThemeStore();
