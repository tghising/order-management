import Vue from 'vue'
import App from './App.vue'
import {BootstrapVue} from 'bootstrap-vue'
import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue/dist/bootstrap-vue.css'
import router from './router'
import DatePicker from "vue2-datepicker";
import 'vue2-datepicker/index.css';

Vue.config.productionTip = false

Vue.use(BootstrapVue)
Vue.use(DatePicker)

new Vue({
  router,
  render: h => h(App),
}).$mount('#app')
