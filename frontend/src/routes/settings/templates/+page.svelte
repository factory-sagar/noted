<script lang="ts">
  import { onMount } from 'svelte';
  import { 
    ArrowLeft,
    FileText,
    Plus,
    Trash2,
    Save,
    Copy
  } from 'lucide-svelte';
  import { addToast } from '$lib/stores';

  interface Template {
    id: string;
    name: string;
    type: 'initial' | 'followup' | 'custom';
    content: string;
    fields: string[];
  }

  let templates: Template[] = [];
  let editingTemplate: Template | null = null;
  let showNewTemplateModal = false;
  let newTemplateName = '';

  const defaultFields = [
    'Account Overview',
    'Attendees',
    'Agenda',
    'Discussion Points',
    'Action Items',
    'Next Steps',
    'Technical Requirements',
    'Timeline',
    'Budget Discussion',
    'Stakeholders'
  ];

  const defaultTemplates: Template[] = [
    {
      id: 'initial',
      name: 'Initial Call',
      type: 'initial',
      content: `<h2>Meeting Overview</h2>
<p>Brief description of the meeting purpose...</p>

<h2>Attendees</h2>
<p>List participants here...</p>

<h2>Company Background</h2>
<p>What does the company do? Industry? Size?</p>

<h2>Current Challenges</h2>
<p>What problems are they trying to solve?</p>

<h2>Technical Requirements</h2>
<ul>
<li>Requirement 1</li>
<li>Requirement 2</li>
</ul>

<h2>Timeline & Budget</h2>
<p>Expected timeline and budget range...</p>

<h2>Next Steps</h2>
<ul>
<li>Action item 1</li>
<li>Action item 2</li>
</ul>`,
      fields: ['Account Overview', 'Attendees', 'Technical Requirements', 'Timeline', 'Budget Discussion', 'Next Steps']
    },
    {
      id: 'followup',
      name: 'Follow-up Call',
      type: 'followup',
      content: `<h2>Meeting Summary</h2>
<p>Quick recap of previous discussions...</p>

<h2>Progress Update</h2>
<p>What has happened since last meeting?</p>

<h2>Discussion Points</h2>
<ul>
<li>Point 1</li>
<li>Point 2</li>
</ul>

<h2>Open Questions</h2>
<p>Questions that need answering...</p>

<h2>Action Items</h2>
<ul>
<li>[ ] Action 1</li>
<li>[ ] Action 2</li>
</ul>`,
      fields: ['Discussion Points', 'Action Items', 'Next Steps']
    }
  ];

  onMount(() => {
    loadTemplates();
  });

  function loadTemplates() {
    const saved = localStorage.getItem('noteTemplates');
    if (saved) {
      templates = JSON.parse(saved);
    } else {
      templates = [...defaultTemplates];
    }
  }

  function saveTemplates() {
    localStorage.setItem('noteTemplates', JSON.stringify(templates));
  }

  function editTemplate(template: Template) {
    editingTemplate = { ...template };
  }

  function saveTemplate() {
    if (!editingTemplate) return;
    
    const index = templates.findIndex(t => t.id === editingTemplate!.id);
    if (index >= 0) {
      templates[index] = editingTemplate;
    } else {
      templates = [...templates, editingTemplate];
    }
    
    saveTemplates();
    editingTemplate = null;
    addToast('success', 'Template saved');
  }

  function deleteTemplate(id: string) {
    if (id === 'initial' || id === 'followup') {
      addToast('error', 'Cannot delete default templates');
      return;
    }
    
    if (!confirm('Delete this template?')) return;
    
    templates = templates.filter(t => t.id !== id);
    saveTemplates();
    addToast('success', 'Template deleted');
  }

  function createTemplate() {
    if (!newTemplateName.trim()) return;
    
    const id = `custom-${Date.now()}`;
    const newTemplate: Template = {
      id,
      name: newTemplateName.trim(),
      type: 'custom',
      content: '<h2>Meeting Notes</h2>\n<p>Start writing here...</p>',
      fields: ['Discussion Points', 'Action Items']
    };
    
    templates = [...templates, newTemplate];
    saveTemplates();
    showNewTemplateModal = false;
    newTemplateName = '';
    editTemplate(newTemplate);
    addToast('success', 'Template created');
  }

  function duplicateTemplate(template: Template) {
    const id = `custom-${Date.now()}`;
    const newTemplate: Template = {
      ...template,
      id,
      name: `${template.name} (Copy)`,
      type: 'custom'
    };
    
    templates = [...templates, newTemplate];
    saveTemplates();
    addToast('success', 'Template duplicated');
  }

  function resetToDefaults() {
    if (!confirm('Reset all templates to defaults? Custom templates will be deleted.')) return;
    
    templates = [...defaultTemplates];
    saveTemplates();
    addToast('success', 'Templates reset to defaults');
  }

  function toggleField(field: string) {
    if (!editingTemplate) return;
    
    if (editingTemplate.fields.includes(field)) {
      editingTemplate.fields = editingTemplate.fields.filter(f => f !== field);
    } else {
      editingTemplate.fields = [...editingTemplate.fields, field];
    }
  }
</script>

<svelte:head>
  <title>Templates - SE Notes</title>
</svelte:head>

<div class="max-w-4xl mx-auto">
  <div class="flex items-center gap-4 mb-8">
    <a href="/settings" class="p-2 hover:bg-[var(--color-card)] rounded-lg transition-colors">
      <ArrowLeft class="w-5 h-5" />
    </a>
    <div class="flex-1">
      <h1 class="page-title">Note Templates</h1>
      <p class="page-subtitle">Customize templates for different meeting types</p>
    </div>
    <button class="btn-secondary" on:click={resetToDefaults}>
      Reset Defaults
    </button>
    <button class="btn-primary" on:click={() => showNewTemplateModal = true}>
      <Plus class="w-4 h-4" />
      New Template
    </button>
  </div>

  {#if editingTemplate}
    <!-- Template Editor -->
    <div class="card">
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-lg font-semibold">Edit Template</h2>
        <div class="flex gap-3">
          <button class="btn-secondary" on:click={() => editingTemplate = null}>
            Cancel
          </button>
          <button class="btn-primary" on:click={saveTemplate}>
            <Save class="w-4 h-4" />
            Save
          </button>
        </div>
      </div>

      <div class="space-y-6">
        <div>
          <label class="label">Template Name</label>
          <input 
            type="text" 
            class="input" 
            bind:value={editingTemplate.name}
            disabled={editingTemplate.type !== 'custom'}
          />
        </div>

        <div>
          <label class="label">Suggested Fields</label>
          <p class="text-sm text-[var(--color-muted)] mb-3">
            Select which fields to include in this template
          </p>
          <div class="flex flex-wrap gap-2">
            {#each defaultFields as field}
              <button
                class="px-3 py-1.5 rounded-lg text-sm border transition-colors"
                class:bg-primary-500={editingTemplate.fields.includes(field)}
                class:text-white={editingTemplate.fields.includes(field)}
                class:border-primary-500={editingTemplate.fields.includes(field)}
                class:border-[var(--color-border)]={!editingTemplate.fields.includes(field)}
                on:click={() => toggleField(field)}
              >
                {field}
              </button>
            {/each}
          </div>
        </div>

        <div>
          <label class="label">Default Content (HTML)</label>
          <textarea 
            class="input font-mono text-sm h-64"
            bind:value={editingTemplate.content}
          ></textarea>
          <p class="text-xs text-[var(--color-muted)] mt-2">
            Use HTML tags like &lt;h2&gt;, &lt;p&gt;, &lt;ul&gt;, &lt;li&gt; for formatting
          </p>
        </div>
      </div>
    </div>
  {:else}
    <!-- Templates List -->
    <div class="grid gap-4">
      {#each templates as template}
        <div class="card">
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <div class="flex items-center gap-3 mb-2">
                <FileText class="w-5 h-5 text-primary-500" />
                <h3 class="font-semibold">{template.name}</h3>
                {#if template.type !== 'custom'}
                  <span class="text-xs px-2 py-0.5 bg-primary-500/10 text-primary-500 rounded">
                    Default
                  </span>
                {/if}
              </div>
              <p class="text-sm text-[var(--color-muted)] mb-3">
                Fields: {template.fields.join(', ')}
              </p>
              <div class="flex gap-2">
                <button 
                  class="btn-secondary text-sm py-1.5"
                  on:click={() => editTemplate(template)}
                >
                  Edit
                </button>
                <button 
                  class="btn-ghost text-sm py-1.5"
                  on:click={() => duplicateTemplate(template)}
                >
                  <Copy class="w-4 h-4" />
                  Duplicate
                </button>
                {#if template.type === 'custom'}
                  <button 
                    class="btn-ghost text-sm py-1.5 text-red-500"
                    on:click={() => deleteTemplate(template.id)}
                  >
                    <Trash2 class="w-4 h-4" />
                    Delete
                  </button>
                {/if}
              </div>
            </div>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- New Template Modal -->
{#if showNewTemplateModal}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
    <button 
      class="absolute inset-0 bg-black/50 backdrop-blur-sm"
      on:click={() => showNewTemplateModal = false}
    ></button>
    <div class="relative bg-[var(--color-card)] border border-[var(--color-border)] rounded-xl p-6 w-full max-w-md animate-slide-up">
      <h2 class="text-lg font-semibold mb-4">New Template</h2>
      <form on:submit|preventDefault={createTemplate}>
        <div class="mb-4">
          <label class="label">Template Name</label>
          <input 
            type="text"
            class="input"
            placeholder="e.g., Technical Deep Dive"
            bind:value={newTemplateName}
            autofocus
          />
        </div>
        <div class="flex justify-end gap-3">
          <button 
            type="button"
            class="btn-secondary"
            on:click={() => showNewTemplateModal = false}
          >
            Cancel
          </button>
          <button type="submit" class="btn-primary" disabled={!newTemplateName.trim()}>
            Create
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}
