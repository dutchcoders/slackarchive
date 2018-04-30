import Vue from 'vue'

// Event proxy/bus for sharing data across non-parent child components
// https://vuejs.org/v2/guide/components.html#Non-Parent-Child-Communication

const event = {
  bus: null,

  init() {
    if (!this.bus) {
      this.bus = new Vue();
    }

    return this;
  },

  emit(name, ...args) {
    this.bus.$emit(name, ...args);
    return this;
  },

  on() {
    if (arguments.length === 2) {
      this.bus.$on(arguments[0], arguments[1]);
    } else {
      Object.keys(arguments[0]).forEach(key => {
        this.bus.$on(key, arguments[0][key]);
      });
    }

    return this;
  },
}

export {event};
