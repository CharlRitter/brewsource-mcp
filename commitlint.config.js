module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    // Type enum - allowed commit types
    'type-enum': [
      2,
      'always',
      [
        'feat',     // New feature
        'fix',      // Bug fix
        'docs',     // Documentation changes
        'style',    // Code style changes (formatting, etc.)
        'refactor', // Code refactoring
        'test',     // Adding or updating tests
        'chore',    // Maintenance tasks
        'perf',     // Performance improvements
        'ci',       // CI/CD changes
        'build',    // Build system changes
        'revert'    // Revert previous commit
      ]
    ],

    // Subject case - sentence case (first letter capitalized)
    'subject-case': [2, 'always', 'sentence-case'],

    // Subject length limits
    'subject-max-length': [2, 'always', 72],
    'subject-min-length': [2, 'always', 10],

    // Body line length
    'body-max-line-length': [2, 'always', 100],

    // Header length
    'header-max-length': [2, 'always', 72],

    // Scope rules
    'scope-case': [2, 'always', 'lower-case'],

    // Subject rules
    'subject-empty': [2, 'never'],
    'subject-full-stop': [2, 'never', '.'],

    // Type rules
    'type-empty': [2, 'never'],
    'type-case': [2, 'always', 'lower-case'],

    // Footer rules for breaking changes and issue references
    'footer-leading-blank': [1, 'always'],
    'footer-max-line-length': [2, 'always', 100],

    // Body rules
    'body-leading-blank': [1, 'always'],

    // Signed-off-by line
    'signed-off-by': [0, 'always', 'Signed-off-by:']
  },

  // Custom validation functions
  plugins: [
    {
      rules: {
        // Custom rule to check for brewing-specific terms
        'brewing-terminology': (parsed) => {
          const { subject } = parsed;
          if (!subject) return [true];

          // Check for common brewing acronyms that should be uppercase
          const brewingTerms = /\b(bjcp|mcp|ibu|abv|srm|og|fg|api|json|http|sql|ui|ux)\b/gi;
          const matches = subject.match(brewingTerms);

          if (matches) {
            const hasLowercase = matches.some(match => match !== match.toUpperCase());
            if (hasLowercase) {
              return [
                false,
                'Brewing acronyms should be uppercase (BJCP, MCP, IBU, ABV, SRM, OG, FG, API, JSON, HTTP, SQL, UI, UX)'
              ];
            }
          }

          return [true];
        }
      }
    }
  ],

  // Ignore certain patterns
  ignores: [
    (message) => message.includes('WIP:'),
    (message) => message.includes('Merge'),
    (message) => message.includes('Revert')
  ],

  // Default severity level
  defaultIgnores: true,

  // Help URL for users
  helpUrl: 'https://github.com/conventional-changelog/commitlint/#what-is-commitlint'
};
