<template>
  <div id="app" :class="{'with-menu': showMenu, 'ready': appIsReady}">
    <div class="splash" v-show="!appIsReady">
      <img src="./assets/logo.png" alt="SlackArchive.io" class="img-fluid">
      <div class="loader loader-sm" v-show="!error"></div>
    </div>
    <div v-if="error" :class="'error-'+errorType" class="error">{{ error }}
      <span v-if="!errorType">
        <span class="smiley">☹</span>
        <a href="http://slackarchive.io/#support" target="_blank">Let us know</a>
      </span>
    </div>
    <top-bar v-if="appIsReady" :team="team" @toggleMenu="onMenuToggle"></top-bar>
    <side-bar v-if="appIsReady" :team="team" :channels="channels"></side-bar>
    <div class="main-content" id="main_container">
      <router-view :channel="channel" :team="team" :channelIDsToName="channelIDsToName" :appIsReady="appIsReady" v-on:error="onError"></router-view>
    </div>
  </div>
</template>

<script>
  import axios from 'axios';
  import TopBar from './components/TopBar'
  import SideBar from './components/SideBar'
  import Services from './services';

  export default {
    name: 'app',
    components: {
      TopBar, SideBar
    },
    data () {
      return {
        appIsReady: null,
        showMenu: false,
        error: null,
        errorType: null,
        domain: null,
        team: null,
        channels: [],
        users: null,
        channel: null,
      }
    },
    created () {
      axios.defaults.headers.common['X-Alt-Referer'] =
        (document.location.origin)
    },
    mounted () {
      this.detectTeam()
    },
    watch: {
      $route(newRoute, oldRoute) {
        if (newRoute.name == 'home') {
          this.channel = null
        } else {
          this.detectChannel()
        }

        if (this.showMenu)
          this.showMenu = false
      }
    },
    computed: {
      channelIDsToName() {
        let channelIDsToName = {}
        this.channels && this.channels.forEach((channel) => {
          channelIDsToName[channel.channel_id] = channel.name
        })
        return channelIDsToName
      }
    },
    methods: {
      detectTeam () {
        this.getTeams(null)
      },
      detectChannel () {
        const channel_name = this.$route.params.channel_name;
        if (!channel_name) {
          const defaultChannel = this.channels.find(each => {
            return each.is_general === true
          });
          if (defaultChannel) {
            this.$router.push({name: 'channel', params: {'channel_name': defaultChannel.name}})
          }
          return;
        }
        this.channel = channel_name === '-' ? {name: '-', channel_id: ''} : this.channels.find(each => {
            return each.name == channel_name
          });
        if (!this.channel && channel_name !== '-') {
          this.error = 'Channel "' + channel_name + '" not found.'
          this.errorType = 'known'
        }
      },
      getTeams (teamDomain) {
        Services.getTeams(teamDomain).then(response => {
          if (!response.data.team) {
            this.error = 'Team by domain "' + teamDomain + '" not found.'
            this.errorType = 'known'
            return
          }
          this.team = response.data.team[0]
          this.getChannels()
        }).catch(this.onError);
      },
      getChannels () {
        Services.getChannels(this.team.team_id).then(response => {
          this.channels = response.data.channels
          if (!this.channels) {
            this.error = 'No channels archived yet for "' + this.team.domain + '" team.'
            this.errorType = 'known'
            return
          }
          this.appIsReady = true
          this.detectChannel()
        }).catch(this.onError);
      },
      onMenuToggle () {
        this.showMenu = !this.showMenu
      },
      onError (error) {
        if (error) {
          this.error = error.response && error.response.data ? error.response.data : 'Unexpected error!'
        } else {
          this.error = null
          this.errorType = null
        }
      }
    }
  }
</script>

<style lang="sass">
  @import "sass/app.scss";
</style>
