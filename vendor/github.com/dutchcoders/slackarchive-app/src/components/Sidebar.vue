<template>
  <div class="sidebar" ng-cloak>
    <router-link :to="{ name: 'home'}" v-if="team" class="team-name">{{ team.name }}</router-link>

    <div class="content">

      <form class="search-form" id="searchForm" @submit.prevent="performSearch">
        <div class="input-group">
          <input type="text" class="search-input" size="30" placeholder="Search..." v-model="search_query">
          <span v-show="!search_query" class="icon icon-search"></span>
          <a href="" @click.prevent="resetSearch"><span v-show="search_query" class="icon icon-remove"></span></a>
        </div>

        <!--<div class="row dates">
          <div class="col-6">
            <div class="input-group date-from">
              <div class="daterange daterange&#45;&#45;single" date-range></div>
            </div>
          </div>
          <div class="col-6">
            <div class="input-group date-to">
              <div class="daterange daterange&#45;&#45;single" date-range></div>
            </div>
          </div>
        </div>-->
      </form>

      <h2 id="sepH2">
      <span class="filters" v-show="!search_query">
        <a href="" v-on:click.prevent="activeFilter = true" :class="{active: activeFilter}">Active</a>
        <a href="" v-on:click.prevent="activeFilter = null" :class="{active: !activeFilter}">All</a>
      </span>
        <span v-show="!areFilteredChannels">Channels</span>
        <span v-show="areFilteredChannels === true">Filter by channel</span>
      </h2>

      <div class="shadow-container" id="shadowContainer">
        <ul class="channels-list" id="channelsList">
          <li v-for="each in filteredChannels">
            <router-link :to="{ name: (search_query ? 'channel_search': 'channel'), params: { channel_name: each.name, search: search_query }}" :id="'channel-menu-'+each.name">
              # {{ each.name }}
              <span class="count" v-if="each.results">{{ each.results }}</span>
            </router-link>
          </li>
        </ul>
        <div class="sofa-scrolling-shadow-top" id="topShadow"></div>
        <div class="sofa-scrolling-shadow-bottom" id="bottomShadow"></div>
      </div>

      <!--<div ng-if="teamsView">
        <h2>Teams</h2>

        <div class="shadow-container">
          <ul class="channels-list" shadow-scroll="channels" ng-cloak>
            <li ng-repeat="team in teams | orderBy:'name'">
              <a ng-href="{{::team.href}}">{{::team.name}}</a>
            </li>
          </ul>
        </div>
      </div>-->

    </div>
  </div>
</template>

<script>
  import {getEl, elHeight, elOffsetTop, elFullHeight, winHeight} from '../utils.js'
  import orderBy from 'lodash/orderBy'
  import throttle from 'lodash/throttle'
  import {event}  from '../event'

  let searchForm, channelsList, sepH2, topShadow, bottomShadow, topShadowHeight, bottomShadowHeight, searchInterval

  event.init()

  export default {
    props: ['team', 'channels'],
    data () {
      return {
        msg: '',
        search: null,
        channelAggs: null,
        areFilteredChannels: null,
        channelAggsForQuery: null,
        search_query: this.$route.params.search_query || null,
        activeFilter: true,
      }
    },
    computed: {
      filteredChannels() {

        let channels = orderBy(this.channels.filter(channel => {
          return !this.activeFilter || (this.activeFilter && !channel.is_archived)
        }), 'name');

        let channelsWithResults = []
        this.areFilteredChannels = false
        if (this.channelAggs && this.search_query && this.search_query == this.channelAggsForQuery) {
          for (const key in this.channelAggs) {
            if (this.channelAggs.hasOwnProperty(key)) {
              this.channels.forEach((channel) => {
                if (channel.channel_id == key) {
                  this.areFilteredChannels = true
                  channel.results = this.channelAggs[key]
                  channelsWithResults.push(channel)
                }
              })
            }
          }
          channels = orderBy(channelsWithResults, 'results', 'desc')
        } else {
          this.channels.forEach((channel) => {
            delete channel.results
          })
        }

        return channels
      }
    },
    watch: {
      $route (route, oldRoute) {
        this.onRouteChange(route, oldRoute)
      }
    },
    mounted () {
      searchForm = getEl('searchForm')
      channelsList = getEl('channelsList')
      sepH2 = getEl('sepH2')
      topShadow = getEl('topShadow')
      bottomShadow = getEl('bottomShadow')

      this.onResize();
      window.addEventListener('resize', throttle(this.onResize, 200), false);

      this.handleShadow()
      this.onRouteChange(this.$route, null, true)

      event.on('channelAggs', (aggs, searchQuery) => {
        this.channelAggs = aggs
        this.channelAggsForQuery = searchQuery
      })
    },

    methods: {
      onRouteChange (route, oldRoute, byMount = false) {
        if (route.name.indexOf('channel_search') === 0) {
          if (byMount || this.search_query !== route.params.search) {
            this.search_query = route.params.search
          }
        } else if (this.search || this.search_query) {
          this.search = this.search_query = null
        }
      },
      performSearch (){
        let newSearch = this.search_query ? {query: this.search_query} : null;
        if (newSearch != this.search)
          this.$router.push({name: 'channel_search', params: {channel_name: '-', search: this.search_query}})
        this.search = newSearch;
      },
      resetSearch () {
        this.search_query = null
        this.search = null
        this.$router.push(this.$route.params.channel_name && this.$route.params.channel_name !== '-' ? {
            name: 'channel',
            params: this.$route.params.channel_name
          } : {name: 'home'})
      },
      onResize () {
        let available = winHeight() - elOffsetTop(searchForm) - elHeight(searchForm) - elFullHeight(sepH2)
        channelsList.style.height = Math.floor(available) + 'px';
      },

      handleShadow () {
        topShadowHeight = topShadow.clientHeight
        bottomShadowHeight = bottomShadow.clientHeight
        this.onChannelsScroll()
        channelsList.addEventListener('scroll', throttle(this.onChannelsScroll, 100), false)
      },

      onChannelsScroll () {
        let scrollTop = channelsList.scrollTop,
          scrollBottom = channelsList.scrollHeight - scrollTop - channelsList.clientHeight,
          rollingShadowOffsetTop = 0,
          rollingShadowOffsetBottom = 0;

        if (scrollTop < topShadowHeight) {
          rollingShadowOffsetTop = (topShadowHeight - scrollTop) * -1;
        }
        if (scrollBottom < bottomShadowHeight) {
          rollingShadowOffsetBottom = (bottomShadowHeight - scrollBottom) * -1;
        }

        topShadow.style.top = rollingShadowOffsetTop + 'px';
        bottomShadow.style.bottom = rollingShadowOffsetBottom + 'px';
      }
    }

  }
</script>
