import { Extension } from '@tiptap/core';
import Suggestion from '@tiptap/suggestion';
import { PluginKey } from 'prosemirror-state';

export const WikiLink = Extension.create({
  name: 'wikiLink',

  addOptions() {
    return {
      suggestion: {
        char: '[[',
        pluginKey: new PluginKey('wikiLinkSuggestion'),
        command: ({ editor, range, props }) => {
          const { id, title, type } = props;
          // Insert link as HTML or custom node. For now simple HTML link
          // Or better: standard link mark with a special class
          // But wait, TipTap link is a mark.
          // Let's insert text and link it.
          
          // We can use the standard Link extension or just text.
          // Let's create a link to the note/account
          const href = type === 'note' ? `/notes/${id}` : `/accounts`; // account linking minimal for now
          
          editor
            .chain()
            .focus()
            .deleteRange(range)
            .setLink({ href })
            .insertContent(title)
            .unsetLink() // Stop linking for next chars
            .run();
            
          // Actually better: use Mention extension logic but with custom renderer
          // But 'Link' mark is simpler.
          // Let's try to just insert a standard link.
          
          editor.chain().focus().deleteRange(range)
            .insertContent([
              {
                type: 'text',
                text: title,
                marks: [
                  {
                    type: 'link',
                    attrs: {
                      href: href,
                      target: '_self', // Internal navigation
                      class: 'wiki-link text-primary-500 hover:underline font-medium'
                    }
                  }
                ]
              },
              {
                type: 'text',
                text: ' ' // Space after
              }
            ])
            .run();
        },
      },
    };
  },

  addProseMirrorPlugins() {
    return [
      Suggestion({
        editor: this.editor,
        ...this.options.suggestion,
      }),
    ];
  },
});
