import { defineConfig } from 'orval';

export default defineConfig({
  portfolio: {
    input: 'http://localhost:8889/openapi.json',
    output: {
      target: './lib/api/portfolio.ts',
      client: 'react-query',
      mode: 'tags-split',
      override: {
        mutator: {
          path: './lib/api/mutator.ts',
          name: 'customInstance',
        },
      },
    },
  },
});
