<template>
  <router-view />
</template>

<script setup>
</script>

<style>
:root {
  --primary-color: #2b2b2b;
  --price-color: #d44e4e;
  --bg-gray: #f7f7f7;
  --border-color: #eeeeee;
  --text-primary: #1A1A1A;
  --text-regular: #4A4A4A;
  --text-secondary: #8C8C8C;
  --color-primary: #1D4ED8;
  --color-success: #16A34A;
  --color-warning: #EA580C;
  --color-danger: #DC2626;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial,
    'Noto Sans', sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol',
    'Noto Color Emoji';
}

#app {
  width: 100%;
  min-height: 100vh;
}
</style>
