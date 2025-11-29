<script lang="ts">
  import { theme, type Theme } from '$lib/stores/theme';
  import { Check } from 'lucide-svelte';

  const themes: { id: Theme; name: string; description: string; colors: string[] }[] = [
    { 
      id: 'modern', 
      name: 'Modern SaaS', 
      description: 'Clean, professional, and familiar. The default look.',
      colors: ['#3b82f6', '#f8fafc', '#0f172a']
    },
    { 
      id: 'minimal', 
      name: 'Soft Minimalist', 
      description: 'Calming, rounded, and spacious. Sage & Stone.',
      colors: ['#3ba381', '#fafaf9', '#1c1917']
    },
    { 
      id: 'retro', 
      name: 'Retro Pop', 
      description: 'Vibrant 80s vibe with brutalist shadows and pinks.',
      colors: ['#d946ef', '#fff1f2', '#171717']
    },
    { 
      id: 'nordic', 
      name: 'Nordic Wood', 
      description: 'Organic, earthy pine and cream. Cozy and serif.',
      colors: ['#547a76', '#f4f1ea', '#2c3333']
    },
    { 
      id: 'cyber', 
      name: 'Cyber Dev', 
      description: 'High contrast, monospaced, and technical. Neon Lime.',
      colors: ['#84cc16', '#050505', '#eeeeee']
    },
    { 
      id: 'monokai', 
      name: 'Monokai IDE', 
      description: 'Code editor aesthetic. Dark grey with vivid accents.',
      colors: ['#f92672', '#272822', '#f8f8f2']
    },
    { 
      id: 'noir', 
      name: 'Editorial Noir', 
      description: 'Classic, sophisticated, and warm. Gold & Serif.',
      colors: ['#c9a87c', '#faf9f6', '#1a1a1a']
    },
    { 
      id: 'dracula', 
      name: 'Dracula', 
      description: 'Popular purple dark theme from code editors.',
      colors: ['#bd93f9', '#282a36', '#f8f8f2']
    },
    { 
      id: 'solarized', 
      name: 'Solarized', 
      description: 'Classic warm/cool developer color scheme.',
      colors: ['#268bd2', '#fdf6e3', '#002b36']
    },
    { 
      id: 'ocean', 
      name: 'Ocean', 
      description: 'Deep blue aquatic vibes. Calm and focused.',
      colors: ['#0891b2', '#ecfeff', '#164e63']
    },
    { 
      id: 'forest', 
      name: 'Forest', 
      description: 'Deep emerald and moss greens. Natural.',
      colors: ['#059669', '#ecfdf5', '#064e3b']
    }
  ];

  function selectTheme(id: Theme) {
    theme.set(id);
  }
</script>

<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
  {#each themes as t}
    <button
      class="flex items-start gap-4 p-4 border rounded-lg text-left transition-all relative overflow-hidden group
        {$theme === t.id 
          ? 'bg-[var(--color-primary-50)] border-[var(--color-accent)] ring-1 ring-[var(--color-accent)]' 
          : 'bg-[var(--color-card)] border-[var(--color-border)] hover:border-[var(--color-muted)]'}"
      on:click={() => selectTheme(t.id)}
    >
      <!-- Color Preview -->
      <div class="flex-shrink-0 flex flex-col gap-1 mt-1">
        <div class="flex gap-1">
          <div class="w-4 h-4 rounded-full border border-black/10" style="background-color: {t.colors[0]}"></div>
          <div class="w-4 h-4 rounded-full border border-black/10" style="background-color: {t.colors[1]}"></div>
        </div>
        <div class="w-9 h-4 rounded-full border border-black/10" style="background-color: {t.colors[2]}"></div>
      </div>

      <div class="flex-1 min-w-0">
        <div class="flex items-center justify-between mb-1">
          <span class="font-medium font-sans text-[var(--color-text)]">
            {t.name}
          </span>
          {#if $theme === t.id}
            <div class="flex items-center justify-center w-5 h-5 rounded-full bg-[var(--color-accent)] text-white">
              <Check class="w-3 h-3" strokeWidth={3} />
            </div>
          {/if}
        </div>
        <p class="text-sm text-[var(--color-text-secondary)] leading-relaxed">
          {t.description}
        </p>
      </div>
    </button>
  {/each}
</div>
