// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import VueRouter from 'vue-router'

import App from './App'
import Home from './components/Home.vue'
import Messages from './components/Messages'


Vue.use(VueRouter)

const routes = [
  {
    path: '/',
    name: 'home',
    component: Home
  },
  {
    path: '/:channel_name',
    name: 'channel',
    component: Messages,
    children: [
      {
        path: 'search-:search',
        name: 'channel_search',
        component: Messages,
        children: [
          {
            path: 'page-:page',
            name: 'channel_search_page',
            component: Messages,
            children: [
              {
                path: 'ts-:id',
                name: 'channel_search_message',
                component: Messages
              }
            ]
          }
        ]
      },
      {
        path: 'page-:page',
        name: 'channel_page',
        component: Messages,
        children: [
          {
            path: 'ts-:id',
            name: 'channel_message',
            component: Messages
          }
        ]
      },
    ]
  }
]

const router = new VueRouter({
  routes: routes,
  mode: (document.location.href.indexOf('hash_mode') !== -1 ? 'hash' : 'history'),
})

/* eslint-disable no-new */
new Vue({
  el: '#app',
  template: '<App/>',
  components: {App},
  router
})
