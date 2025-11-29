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
      id: 'cyber', 
      name: 'Cyber Dev', 
      description: 'High contrast, monospaced, and technical. Neon Lime.',
      colors: ['#84cc16', '#050505', '#eeeeee']
    },
    { 
      id: 'noir', 
      name: 'Editorial Noir', 
      description: 'Classic, sophisticated, and warm. Gold & Serif.',
      colors: ['#c9a87c', '#faf9f6', '#1a1a1a']
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
