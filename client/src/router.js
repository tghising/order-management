import Vue from "vue";
import Router from "vue-router";

Vue.use(Router);

export default new Router({
    mode: "history",
    routes: [
        {
            path: "/",
            alias: "/home",
            name: "home",
            component: () => import("./components/Home")
        },
        {
            path: "/about",
            alias: "/about",
            name: "about",
            component: () => import("./components/About")
        },
        {
            path: "/customers",
            alias: "/customers",
            name: "customers",
            component: () => import("./components/Customers")
        },
        {
            path: "/home",
            alias: "/home",
            name: "home",
            component: () => import("./components/Home")
        },
        {
            path: "/orders",
            alias: "/orders",
            name: "orders",
            component: () => import("./components/Orders")
        }
    ]
});
