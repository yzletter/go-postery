module.exports = {
  root: true,
  env: { browser: true, es2020: true },
  parser: '@typescript-eslint/parser',
  extends: [
    'eslint:recommended',
    'plugin:@typescript-eslint/recommended',
    'plugin:react-hooks/recommended',
  ],
  plugins: [],
  ignorePatterns: ['dist', 'node_modules'],
  rules: {
    '@typescript-eslint/no-explicit-any': 'off',
  },
}
