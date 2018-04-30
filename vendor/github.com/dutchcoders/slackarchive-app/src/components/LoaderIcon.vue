<template>
  <div v-show="loading" class="loader" :class="'loader-' + (size ? size : 'xs')">{{size}}</div>
</template>

<script>
  export default {
    props: ['loading', 'size']
  }
</script>
