export const isDevDomain = document.location.host.indexOf('localhost') !== -1

export const getEl = id => {
  return document.getElementById(id)
}

export const elOffsetTop = el => {
  let r = el.getBoundingClientRect();
  return r.top
}

export const elHeight = el => {
  let r = el.getBoundingClientRect();
  return r.bottom - r.top
}

export const elFullHeight = el => {
  let styles = window.getComputedStyle(el),
    margin = parseFloat(styles['marginTop']) +
      parseFloat(styles['marginBottom']);

  return Math.ceil(el.offsetHeight + margin);
}

export const winHeight = () => {
  let w = window,
    d = document,
    e = d.documentElement,
    g = d.getElementsByTagName('body')[0]

  return w.innerHeight || e.clientHeight || g.clientHeight
}

const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

const monthFullNames = ["January", "February", "March", "April", "May", "June",
  "July", "August", "September", "October", "November", "December"]

export const formatDate = (dt, full = false) => {
  if (full)
    return monthFullNames[dt.getMonth()] + ' ' + dt.getDate() + ', ' + dt.getFullYear()

  let hours = dt.getHours(), minutes = dt.getMinutes()
  return monthNames[dt.getMonth()] + ' ' + dt.getDate() + ', ' +
    dt.getFullYear() + ' ' +
    (hours < 10 ? '0' : '') + hours + ':' +
    (minutes < 10 ? '0' : '') + minutes
}

const requestAnimFrame = (function () {
  return window.requestAnimationFrame ||
    window.webkitRequestAnimationFrame ||
    window.mozRequestAnimationFrame ||
    function (callback) {
      window.setTimeout(callback, 1000 / 60);
    };
})();

export const scrollTo = (element, scrollTargetY, speed = 2000, easing = 'easeInOutQuint') => {
  // speed: time in pixels per second

  let scrollY = element.scrollTop,
    currentTime = 0;

  // min time .1, max time .8 seconds
  let time = Math.max(.1, Math.min(Math.abs(scrollY - scrollTargetY) / speed, .4));

  // easing equations from https://github.com/danro/easing-js/blob/master/easing.js
  let easingEquations = {
    easeOutSine: function (pos) {
      return Math.sin(pos * (Math.PI / 2));
    },
    easeInOutSine: function (pos) {
      return (-0.5 * (Math.cos(Math.PI * pos) - 1));
    },
    easeInOutQuint: function (pos) {
      if ((pos /= 0.5) < 1) {
        return 0.5 * Math.pow(pos, 5);
      }
      return 0.5 * (Math.pow((pos - 2), 5) + 2);
    }
  };

  // add animation loop
  function tick() {
    currentTime += 1 / 60;

    let p = currentTime / time,
      t = easingEquations[easing](p);

    if (p < 1) {
      requestAnimFrame(tick);
      element.scrollTop = scrollY + ((scrollTargetY - scrollY) * t);
    } else {
      element.scrollTop = scrollTargetY;
    }
  }

  // call it once to get started
  tick();
}


export const queryParams = (url) => {
  let vars = [], hash, index = url.indexOf('?');
  if (index < 0) {
    return [];
  } else {
    let hashes = url.slice(index + 1).split('&'), i;
    for (i = 0; i < hashes.length; i++) {
      hash = hashes[i].split('=');
      vars.push(hash[0]);
      vars[hash[0]] = hash[1];
    }
    return vars;
  }
}
