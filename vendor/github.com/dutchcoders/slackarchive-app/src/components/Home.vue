<template>
  <div>
    <div v-if="appIsReady">Choose channel</div>
  </div>
</template>

<script>

  export default {
    props: ['appIsReady'],
    watch: {
      channel () {
        this.getMessages()
      }
    },
    mounted () {
//      this.getMessages()
    },
    methods: {
    }
  }
</script>
