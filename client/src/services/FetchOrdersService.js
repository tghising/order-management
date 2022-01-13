import http from "../http-common";

class FetchOrdersService {
    getAll(params) {
        return http.get("/orders", {params});
    }

}

export default new FetchOrdersService();
