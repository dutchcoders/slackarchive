<template>
  <div>

    <loader-icon :loading="isLoading === 'top'" class="messages-loader-top"></loader-icon>

    <ul class="messages" id="messages" :class="{loading: isLoading === true}">

      <li v-if="realTotalMessages > 10000 && currentPage === 1" class="msg-danger">
        Due to performance limitations only 10000 messages can be paginated. <br class="visible-lg"/>
        Use search to filter all messages
        <span v-if="channel.channel_id">in this channel</span> ({{ realTotalMessages }} messages).
      </li>
      <li v-for="m in messages" :class="['msg-type-' + m.subtype, {'msg-same-user': m.same_user, active: activeMessageId == m.ts_id}]" :id="m.ts_id">
        <div class="msg-date-separator" v-show="m.date_separator">
          <span>{{ m.date_separator }}</span>
        </div>
        <div class="msg-avatar">
          <span v-if="m.user.profile && !m.same_user">
            <img :src="m.user.profile.image_48" class="msg-thumb">
          </span>
        </div>
        <div class="msg-container">
          <div class="msg-header" v-show="!m.same_user">
            <span class="msg-user">{{m.user.name}}</span>
            <span class="msg-time">
              <a href="" @click.stop.prevent="getMessageUrl(m)" rel="nofollow">{{ m.date_str }}</a>
              <loader-icon v-if="search && targetMessageIdLoading == m.ts_id" :loading="1" class="messages-loader-msg"></loader-icon>
              <span class="msg-loading-error" v-if="targetMessageError && targetMessageId == m.ts_id">{{ targetMessageError }}</span>
            <br></span>
          </div>
          <div class="msg-body" v-html="m.text"></div>
        </div>
        <div class="msg-date-separator msg-end-separator" v-if="m.isLast">
          <span>༄ No more new messages &nbsp;<em dir="rtl" class="end-mark">༄</em></span>
        </div>
      </li>
      <li v-if="totalMessages === 0 && channel">
        <div class="msg-date-separator msg-end-separator">
          <span><em>༄</em> No messages in channel #{{ channel.name }}&nbsp;<em dir="rtl" class="end-mark">༄</em></span>
        </div>
      </li>
    </ul>

    <loader-icon :loading="isLoading === true" class="messages-loader-center" :size="'sm'"></loader-icon>

    <div class="messages-loader-bottom">
      <loader-icon :loading="isLoading === 'bottom'"></loader-icon>
    </div>

    <div class="pagination-container" v-if="appIsReady" id="pagination">
      <paginate :center-side="2" :side="1" :last-page="lastPage" :current-page="currentPage"
                :css-class="'pagination-vertical'" :next-text="'▼'" :prev-text="'▲'"
                :url-sync="false" @turn="changePage"></paginate>
    </div>

  </div>
</template>

<script>
  import Vue from 'vue'
  import throttle from 'lodash/throttle'
  import isEqual from 'lodash/isEqual'
  import Services from '../services';
  import emoji from 'node-emoji'
  import {formatDate, getEl, winHeight, elOffsetTop, elFullHeight, elHeight, scrollTo} from '../utils'
  import Paginate from './Paginate'
  import LoaderIcon from './LoaderIcon'
  import {event}  from '../event'

  const perPage = 100, keepPages = 4, maxOffset = 10000

  let container, messagesEl, scrollPages, currentWinHeight = winHeight(), scrollThrottle, resizeThrottle, mounted,
    channelAggs

  event.init()

  export default {
    props: ['channel', 'team', 'channelIDsToName', 'appIsReady'],

    components: {
      Paginate, LoaderIcon
    },

    data () {
      return {
        messages: [],
        users: [],

        search: null,
        componentIsReady: null,
        activeMessageId: null,
        targetMessageId: null,
        targetMessageIdLoading: null,
        targetMessageError: null,
        error: null,
        isLoading: null,
        realTotalMessages: null,
        totalMessages: null,
        lastPage: 0,
        currentPage: 0
      }
    },

    watch: {
      error (error) {
        this.$emit('error', error)
      },
      appIsReady (isReady) {

        // Initial load
        if (!mounted)
          this.onMounted()

        Vue.nextTick(() => {
          if (isReady) {
            // Move pagination inside main container, or it is getting hidden in Safari because of
            // main container being fixed too and having css right param
            let pagination = getEl('pagination');
            pagination && getEl('app').appendChild(pagination)
          }
        })
      },
      $route (route, oldRoute) {
        this.onRouteChange(route, oldRoute)
      }
    },

    mounted () {
      if (!this.appIsReady)
        return
      this.onMounted()
    },

    beforeDestroy () {
      this.stopTrackScroll();
      this.stopTrackResize();
      this.reset()
      this.emitChannelAggs(null)
      let pagination = getEl('pagination');
      pagination && getEl('app').removeChild(getEl('pagination'))
      mounted = false
    },

    methods: {
      reset() {
        this.totalMessages = null
        this.realTotalMessages = null
        this.currentPage = 0
        this.lastPage = 0
        this.error = null
        this.search = null
        this.targetMessageId = null
        this.targetMessageIdLoading = null
        this.targetMessageError = null
        scrollPages = {}
      },

      onMounted () {
        mounted = true
        container = getEl('main_container');
        messagesEl = getEl('messages')
        this.onRouteChange(this.$route, null, true)
      },

      onRouteChange (route, oldRoute, byMount = false) {
        // byMount is performed on initial load/reload, or messages component being mounted

        let reloadNeeded = true
        console.log('Route change' + (byMount ? ' byMount' : ''), route.name, route.params)

        let messageFound = this.markMessage()

        if (!byMount && (
            (
              messageFound
              // Do not reload if same page (anchoring through message)
//              (route.name === 'channel_message' || route.name === 'channel_search_message')
//              && this.channel.name === route.params.channel_name
//              && this.currentPage === parseInt(route.params.page)
            ) || (
              // Do not reload while paginating, infinite scroll is extending messages array by itself
              // Paginator sets number, while from location comes string (little hackish, but works perfect)
              route.name === 'channel_page' && this.channel.name === route.params.channel_name
            ) || (
              // Search pagination
              route.name === 'channel_search_page'
              && this.channel.name === oldRoute.params.channel_name
              && this.search && this.search.query === route.params.search
            )
          )) {
          reloadNeeded = false
        }

        // Reset if route changed completely
        if (reloadNeeded === false
          && this.channel.name !== oldRoute.params.channel_name
//          && route.name !== 'channel_message' && route.name !== 'channel_search_message'
//          && route.name.replace('_page', '') !== oldRoute.name.replace('_page', '')
        ) {
          this.reset();
        }

        if (reloadNeeded)
          this.initMessages()

        if (byMount)
          return


        // Pagination change
        let newPage = parseInt(route.params.page)
        if (newPage && newPage != this.currentPage) {
          this.changePage(newPage)
        }
      },

      onError (error) {
        this.error = error
      },

      changePage: function (page, extend) {
        if (this.isLoading)
          return

        if (!extend) {
          if (scrollPages[page] !== undefined) {
            this.scrollToEl(getEl(scrollPages[page].first))
            return
          }
          // Reset scroll cache if manually clicked on navigation. Auto load is only used on scroll.
          scrollPages = {}
        }

        this.getMessages(extend, page)
      },

      setCurrentPage (page, setUrl = true, replaceUrl = true) {
        if (page == this.currentPage)
          return

        //console.log('Changing page from', this.currentPage, 'to', page, setUrl, replaceUrl)
        this.currentPage = page

        if (setUrl && this.$route.params.page != page) {
          //console.log('Changing url to page', page, setUrl, replaceUrl)
          this.setPageUrl(replaceUrl)
        }
      },

      setPageUrl (replaceUrl = true) {
        if (!this.currentPage)
          return

        let func = replaceUrl ? this.$router.replace : this.$router.push
        let routeName = this.$route.name, params = {...this.$route.params}
        if (routeName === 'channel_message' || routeName === 'channel_search_message') {
          routeName = 'channel_page'
        } else if (routeName.indexOf('_page') === -1) {
          routeName += '_page'
        }
        params.page = this.currentPage
        delete params.id

        func.call(this.$router, {name: routeName, params: params})
      },

      markMessage(animate = true) {
        this.activeMessageId = null
        if (this.currentPage === 0)
          return
        let newMessageId = this.$route.params.id || null;
        if (newMessageId != this.activeMessageId) {
          if (newMessageId) {
            let messageEl = getEl(newMessageId);
            if (messageEl) {
              this.activeMessageId = newMessageId
              this.scrollToEl(messageEl, animate, -15) // -10 to prevent infinite scroll call
              return true
            } else {
              console.warn('Message with ID: ' + newMessageId + ' not found on page: ' + this.currentPage)
            }
          }
        }
      },

      handleSearch () {
        let searchQuery = this.$route.params.search;
        this.search = searchQuery ? {query: searchQuery} : null
      },

      initMessages () {
        this.reset()
        this.handleSearch()
        this.getMessages(null, parseInt(this.$route.params.page))
      },

      getMessages (extend = null, page, channelId, messageTs, messagesBeforeTs) {
        if (!messageTs)
          this.isLoading = !extend ? true : extend
        page = page || 0

        // Perform request to get total number of messages and calculate offset properly
//        console.log(page, this.lastPage, this.$route.params)
        if (page > 0 && this.lastPage === 0) {
          channelId = channelId ? channelId : this.channel.channel_id;
          Services.getMessages(this.team.team_id, channelId, 0, undefined, !messageTs ? this.search : null)
            .then(response => {
              this.setCount(response.data.total)
              this.calculatePages(page)
              if (!messageTs) {
                this._getMessages(extend, page);
              }
              else {
                // Special case to redirect to exact page where messages is
                // Now we know how many messages are before target message (messagesBeforeTs) and total
                this.targetMessageIdLoading = null
                let messagesAfter = this.realTotalMessages - messagesBeforeTs
                if (messagesAfter > maxOffset) {
                  this.targetMessageError = 'Message context is unreachable'
                  return
                }

                this.search = null
                page = this.lastPage - Math.floor(messagesAfter / perPage)
//                console.log('Messages before:', messagesBeforeTs, 'Total:', this.realTotalMessages, 'Page:', page)

                // Redirect to message
                this.$router.push({
                  name: 'channel_message',
                  params: {
                    channel_name: this.channelIDsToName[channelId],
                    page: page,
                    id: messageTs
                  }
                })

              }
            }).catch(this.onError);
          return
        }

        this._getMessages(extend, page);
      },

      _getMessages: function (extend, page) {
        let offset = perPage * Math.max(0, this.lastPage - page);
        Services.getMessages(this.team.team_id, this.channel.channel_id, perPage, offset, this.search).then(response => {
          this.onMessagesData(response.data, extend, page)
        }).catch(this.onError);
      },

      setCount: function (total) {
        this.realTotalMessages = total
        this.totalMessages = Math.min(total, maxOffset)
      },

      onMessagesData (data, extend, page) {
        this.users = data.related.users
        this.setCount(data.total);

        // Dispatch aggs when search is for all channels
        if (this.search && !isEqual(channelAggs, data.aggs.buckets)) {
          // Todo: make for data.aggs single source of data, using bus vue
          this.emitChannelAggs(data.aggs.buckets);
        }

        // Calculate totals and correct current page for initial list
        page = this.calculatePages(page)

        if (!extend) {
          // Update current page on on manual pagination only
          // For infinite scroll currentPage is updated based on scroll position, to avoid multiple changes
          this.setCurrentPage(page, true, !page)
        }

        let newMessages = this.formatMessages(data.messages, page);
        if (extend === null) {
          this.messages = newMessages
        } else if (extend === 'top') {
          this.messages = [...newMessages, ...this.messages.slice(0, (keepPages - 1) * perPage)];
        } else if (extend === 'bottom') {
          this.messages = [...this.messages.slice(-(keepPages - 1) * perPage), ...newMessages];
        }

        if (!this.messages.length) {
          this.isLoading = false
          return
        }

        // Save first and last message id of each page for scrolling
        scrollPages[page] = {
          first: data.messages[0].ts_id,
          last: data.messages[data.messages.length - 1].ts_id
        }

        // Cleanup scrollPages based on how many pages are kept
        this.iterateScrollPages((key) => {
          if ((extend === 'top' && key > page + keepPages - 1) ||
            (extend === 'bottom' && key < page - keepPages + 1)) {
            delete scrollPages[key]
          }
        })

        // Init infinite pagination both sides
        this.trackScroll()

        // Track window height changes for top/bot detection
        this.trackResize()

        // Auto scroll to end or proper offset for top scroll
        this.handleScroll(extend, page)

        this.isLoading = false
      },

      emitChannelAggs: function (aggs) {
        channelAggs = aggs
        event.emit('channelAggs', channelAggs, this.search ? this.search.query : null)
      },


      iterateScrollPages (func) {
        for (let key in scrollPages) {
          if (!scrollPages.hasOwnProperty(key)) continue
          key = parseInt(key)
          let ret = func(key, scrollPages[key])
          if (ret === false)
            break
        }
      },

      calculatePages (page) {
        const totalPages = Math.ceil(this.totalMessages / perPage)
        this.lastPage = totalPages
        if (page === 0) {
          this.setCurrentPage(totalPages)
          page = this.currentPage
        }

        return page
      },

      handleScroll (extendMode, page) {
        Vue.nextTick(() => {

          let markedEl = this.markMessage(false)

          if (!extendMode && !markedEl) {
            // -1 to prevent scroll to bottom fire
            this.scrollEl(container, container.scrollHeight - currentWinHeight + elOffsetTop(container) - 1)
          }
          else if (extendMode === 'top') {
            let firstEl = getEl(scrollPages[page + 1].first);
            this.scrollToEl(firstEl, false) // Scroll to exact same position
            this.scrollToEl(firstEl, true, -100) // Scroll a little to indicate new content has loaded
          } else if (extendMode === 'bottom') {
            let prevPage = scrollPages[page - 1]
            if (prevPage !== undefined) {
              let lastEl = getEl(scrollPages[page - 1].last);
              let newScrollPos = lastEl.offsetTop - currentWinHeight + elOffsetTop(container) + elFullHeight(lastEl) + 30;
              this.scrollEl(container, newScrollPos, false)
              this.scrollEl(container, newScrollPos, true, 100)
            }
          }

          // Save first message offset positions for scroll spy
          this.iterateScrollPages((key, scrollPage) => {
            let firstEl = getEl(scrollPage.first);
            if (firstEl)
              scrollPage.firstOffsetTop = firstEl.offsetTop
            else
              console.log('Warning: Missing firstEl #1', key, scrollPage)
          })

          this.onScroll(null, page)

        })
      },

      scrollEl (el, to, animate = false, offset = 0) {
        if (animate) {
          return scrollTo(el, to + offset)
        }
        return el.scrollTop = to + offset
      },

      scrollToEl (el, animate = true, offset = 0) {
        if (animate)
          return scrollTo(container, el.offsetTop + offset)
        return container.scrollTop = el.offsetTop + offset
      },

      onScroll (e, page) {
        page = page || this.currentPage
        if (container.scrollTop == container.scrollHeight - currentWinHeight + elOffsetTop(container)) {
          this.onBottomScrollReach(page)
        } else if (container.scrollTop == 0) {
          this.onTopScrollReach(page)
        }

        // Spy scroll for cached pages
        let prevDiff = null, activePage = null
        this.iterateScrollPages((key, scrollPage) => {
          let diff = container.scrollTop - scrollPage.firstOffsetTop
          if (diff >= 0 && (prevDiff === null || diff < prevDiff)) {
            activePage = key
          }
          prevDiff = diff
        })
        if (activePage)
          this.setCurrentPage(activePage)
      },

      onTopScrollReach(page) {
        if (page < 2)
          return
        this.changePage(page - 1, 'top')
      },

      onBottomScrollReach(page) {
        if (page === this.lastPage)
          return
        this.changePage(page + 1, 'bottom')
      },

      trackScroll () {
        if (scrollThrottle)
          return
        scrollThrottle = throttle(this.onScroll, 100)
        container.addEventListener('scroll', scrollThrottle);
      },

      trackResize () {
        if (resizeThrottle)
          return
        resizeThrottle = throttle(this.onResize, 250)
        window.addEventListener('resize', resizeThrottle);
      },

      onResize () {
        currentWinHeight = winHeight()
      },

      stopTrackScroll () {
        if (!scrollThrottle)
          return
        scrollThrottle.cancel()
        container.removeEventListener('scroll', scrollThrottle);
        scrollThrottle = null
      },

      stopTrackResize () {
        if (!resizeThrottle)
          return
        resizeThrottle.cancel()
        window.removeEventListener('resize', resizeThrottle);
        resizeThrottle = null
      },

      getMessageUrl (m){
        if (!this.search) {
          let routeObj = {
            name: 'channel_message',
            params: {channel_name: this.channel.name, page: m.page, id: m.ts_id}
          }
          this.$router.push(routeObj)
          return
        }

        // From search
        this.targetMessageId = m.ts_id
        this.targetMessageIdLoading = m.ts_id
        this.targetMessageError = null
        let channelName = this.channelIDsToName[m.channel];
        if (!channelName) {
          this.targetMessageIdLoading = null
          this.targetMessageError = 'Message channel is not available anymore (ID: ' + m.channel + ')'
          return
        }

        Services.getMessages(this.team.team_id, m.channel, 0, undefined, null, null, m.ts_id)
          .then(response => {
            this.lastPage = 0
            this.currentPage = 0
            let messagesBefore = response.data.total;
            this.getMessages(null, 1, m.channel, m.ts_id, messagesBefore)
          }).catch(this.onError);

      },

      formatMessages (messages, page) {
        if (!messages.length)
          return []

        messages.reverse()
        let msg, i, prev, total = messages.length;

        for (i = 0; i < total; i++) {
          msg = messages[i];

          msg.page = page
          msg.subtype = msg.subtype || ''
          this.formatMessageUser(msg);
          this.formatMessageDate(msg);
          this.formatMessageText(msg.text, msg);
          this.formatMessageAttachments(msg);
          this.formatMessageDateSeparators(msg, prev, page == this.lastPage && i == total - 1);
          prev = msg;

        }
        return messages
      },

      formatMessageDate (msg) {
        msg.ts_id = msg.ts.replace('.', '');
        msg.date = new Date(parseInt(msg.ts * 1000));
        msg.date_str = formatDate(msg.date)
      },

      formatMessageText (text, msg = null) {
        if (!text)
          return

        // Mentions
        text = text.replace(/<(.*?)>/g, (match, contents) => {
            let url, url_name, user;
            if (contents.indexOf('|') !== -1) {
              url = contents.split('|');
              url_name = url[1];
              url = url[0];
            } else {
              url = url_name = contents;
              //var max_len = 50;
              //url_name = $filter('limitTo')(contents, max_len) +
              // (contents.length > max_len + 2 ? '...': '');
            }

            // Replace user name
            if (url_name.indexOf('@') === 0) {
              user = this.users[url_name.substring(1)];
              if (!user) {
                //console.info('User not found', url_name);
                return url_name;
              }
              url_name = '@' + user.name;
            }
            return '<a href="' + url + '" target="_bank">' + url_name + '</a>';
          }
        );

        text = text.replace(/\[hl\](.*?)\[\/hl\]/g, function (match, contents) {
          return '<em class="hl">' + contents + '</em>';
        });

        text = text.replace(/```([\s\S]*?)```/g, function (match, contents) {
          return '<pre>' + contents + '</pre>';
        });

        text = text.replace(/`([\s\S]*?)`/g, function (match, contents) {
          return '<code>' + contents + '</code>';
        });

        // Emoji
        text = emoji.emojify(text)

        if (msg)
          msg.text = text

        return text
      },

      formatMessageUser (msg) {
        let user = this.users[msg.user],
          defaultAvatar = {
            image_48: 'https://i1.wp.com/slack.global.ssl.fastly.net/66f9/img/avatars/ava_0010-48.png'
          }, username;
        if (msg.username) {
          username = this.formatMessageUsername(msg.username);
          msg.user = {
            name: username,
            firstLetter: username.substring(0, 1)
          };
        } else if (msg.user && user) {
          msg.user = user;
        } else if (msg.user && !user) {
          msg.user = {name: 'Unknown'};
        } else if (msg.bot_id) {
          msg.user = {name: 'bot'};
        } else {
          msg.user = {name: 'Unknown'};
        }
        if (!msg.user.profile) {
          msg.user.profile = defaultAvatar
        }
      },

      formatMessageUsername (username) {
        username = username.replace(/<(.*?)>/g, function (match, contents) {
          if (contents.indexOf('|') !== -1) {
            if (contents.indexOf('|') !== -1) {
              return contents.split('|')[1];
            }
          }
        });
        return username;
      },

      formatMessageAttachments (msg) {
        if (!msg.attachments || !msg.attachments.length)
          return;
        msg.text = msg.text || '';
        let att_text = '', i, len = msg.attachments.length, att, text;
        for (i = 0; i < len; i++) {
          att = msg.attachments[i];
          text = att.text || att.fallback;
          text = this.formatMessageText(text);
          att_text += '<div class="msg-attachment" style="border-left-color: #' + att.color + '">' + text + '</div>'
        }
        msg.text = msg.text + att_text;
      },

      formatMessageDateSeparators (msg, prev, isLast) {
        msg.isLast = isLast

        if (!prev || msg.date.getDate() !== prev.date.getDate()) {
          msg.date_separator = formatDate(msg.date, true)
        }

        if (!prev)
          return

        if (!msg.date_separator &&
          msg.user.profile && prev.user.profile &&
          msg.user.profile.image_48 == prev.user.profile.image_48 &&
          msg.user.name == prev.user.name) {

          // Only if delay between messages is no larger than 5min
          if (prev.date - msg.date < 5 * 60 * 1000) {
            msg.same_user = true
          }
        }
      }
    }
  }
</script>
