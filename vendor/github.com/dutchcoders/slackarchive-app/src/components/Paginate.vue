<template>
  <ul v-if="lastPage" class="pagination" :class="cssClass">
    <li class="page-item" v-if="activePrev"><a @click.prevent="prev()" href="" class="page-link">{{ prevText }}</a></li>
    <li v-else class="page-item disabled"><span class="page-link">{{ prevText }}</span></li>
    <li v-for="n in headEnd" class="page-item" :class="{active: isActive(n)}">
      <span v-if="isActive(n)" class="page-link">{{ n }}</span>
      <a v-else @click.prevent="turn(n)" href="" class="page-link">{{ n }}</a>
    </li>
    <li v-if="headEllipsis" class="page-item disabled"><span class="page-link">...</span></li>
    <li v-for="n in centerCount" class="page-item" :class="{active: isActive(n + centerOffset)}">
      <span v-if="isActive(n + centerOffset)" class="page-link">{{ n + centerOffset }}</span>
      <a v-else @click.prevent="turn(n + centerOffset)" href="" class="page-link">{{ n + centerOffset }}</a>
    </li>
    <li v-if="tailEllipsis" class="page-item disabled"><span class="page-link">...</span></li>
    <li v-for="n in tailShowCount" class="page-item" :class="{active: isActive(n + tailOffset)}">
      <span v-if="isActive(n + tailOffset)" class="page-link">{{ n + tailOffset }}</span>
      <a v-else @click.prevent="turn(n + tailOffset)" href="" class="page-link">{{ n + tailOffset }}</a>
    </li>
    <li class="page-item" v-if="activeNext"><a @click.prevent="next()" href="" class="page-link">{{ nextText }}</a></li>
    <li v-else class="page-item disabled"><span class="page-link">{{ nextText }}</span></li>
  </ul>
</template>


<script>
  import Vue from 'vue'
  import {queryParams} from '../utils'

  export default {
    data () {
      return {
        'popStateListening': false
      }
    },
    methods: {
      turn: function (page) {
        if (this.urlSync) {
          let url = window.location.href,
            params = queryParams(url),
            state = {'page': page};
          if (!this.popStateListening) {
            this.popStateListening = true;
            window.addEventListener("popstate", () => {
              this.$emit('turn', window.history.state.page);
            });
          }
          if (!window.history.state) {
            window.history.replaceState({'page': this.currentPage}, '', url);
          }
          if (!params.length) {
            window.history.pushState(state, '', url + '?' + this.pageName + '=' + page);
          } else if (params[this.pageName]) {
            window.history.pushState(state, '', url.replace(new RegExp('(' + this.pageName + '=)\\d+'), '$1' + page));
          } else {
            window.history.pushState(state, '', url + '&' + this.pageName + '=' + page);
          }
        }
        this.$emit('turn', page);
      },
      prev: function () {
        this.turn(this.currentPage - 1);
      },
      next: function () {
        this.turn(this.currentPage + 1);
      },
      isActive: function (n) {
        return this.currentPage == n;
      }
    },
    props: {
      side: {
        type: Number,
        default: 1
      },
      centerSide: {
        type: Number,
        default: 1
      },
      lastPage: {
        type: Number,
        required: true
      },
      currentPage: {
        type: Number,
        required: true
      },
      pageName: {
        type: String,
        default: 'page'
      },
      prevText: {
        type: String,
        default: '«'
      },
      nextText: {
        type: String,
        default: '»'
      },
      urlSync: {
        type: Boolean,
        default: false
      },
      cssClass: {
        type: String,
        default: null
      }
    },
    computed: {
      showCount () {
        return Math.min(this.lastPage, (this.centerSide + this.side) * 2 + 1);
      },
      halfCenterCount() {
        return parseInt((this.showCount - this.side * 2) / 2);
      },
      headEnd () {
        return Math.min(this.lastPage, this.side);
      },
      tailShowCount () {
        return Math.min(this.side, Math.max(this.lastPage - this.side, 0));
      },
      tailOffset () {
        return this.lastPage - this.tailShowCount;
      },
      headEllipsis () {
        return this.lastPage > this.showCount && this.currentPage > this.side + this.halfCenterCount + 1;
      },
      tailEllipsis () {
        return this.lastPage > this.showCount && this.currentPage < this.lastPage - this.halfCenterCount - this.side;
      },
      centerCount () {
        return this.lastPage > this.side * 2
          ? this.lastPage <= this.showCount
            ? this.lastPage - this.side * 2
            : this.showCount - this.side * 2
          : 0;
      },
      centerOffset () {
        return !this.tailEllipsis
          ? this.lastPage - this.showCount + this.side
          : this.headEllipsis
            ? this.currentPage - this.halfCenterCount - 1
            : this.side;
      },
      activePrev () {
        return this.currentPage > 1;
      },
      activeNext () {
        return this.currentPage < this.lastPage;
      }
    },
  }
</script>
