/** @type {import('eslint').FlatConfig[]} */
module.exports = [
  {
    files: ['**/*.js'],
    languageOptions: {
      ecmaVersion: 2021,
      sourceType: 'module',
      globals: {
        window: 'readonly',
        document: 'readonly',
        console: 'readonly',
      },
    },
    rules: {
      quotes: ['error', 'double'],
      semi: ['error', 'always'],
      'no-trailing-spaces': 'error',
      'eol-last': ['error', 'always'],
    },
  },
  {
    files: ['**/*.html'],
    plugins: {
      html: require('eslint-plugin-html'),
    },
    languageOptions: {
      ecmaVersion: 2021,
      sourceType: 'module',
    },
    rules: {
      'no-unused-vars': 'off',
      'no-undef': 'off',
      quotes: ['error', 'double'],
      semi: ['error', 'always'],
      'no-trailing-spaces': 'error',
      'eol-last': ['error', 'always'],
    },
  },
];
