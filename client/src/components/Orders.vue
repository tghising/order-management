<template>
  <div class="list row">
    <div class="col-md-8">
      <form class="d-flex">
        <input class="form-control me-2" type="search" placeholder="Search order by part of order or product name"
               aria-label="Search"
               v-model="searchPartOrderOrProductName">
        <button class="btn btn-outline-success" type="button" @click="page = 1; getOrders();">Search</button>
      </form>
    </div>
    <div class="col-md-8">
      <br/>
      <template>
        <date-picker v-model="orderDateRange" placeholder="Order created date range" range format="YYYY-MM-DD"
                     confirm></date-picker>
      </template>
    </div>

    <div class="col-md-12">
      <hr>
      <p>Total amount: <b> {{ formatPrice(this.grand_total_amount) }}</b></p>
      <b-table :items="orders" :fields="fields" :per-page="0" :current-page="page" small borderless responsive>

        <template #cell(order_name)="data">
          <div class="seprateOrderProduct">{{ data.value }}</div>
        </template>
        <template #cell(order_date)="data">
          {{ formatOrderDate(data.value) }}
        </template>

        <template #cell(delivered_amount)="data">
          {{ formatPrice(data.value) }}
        </template>

        <template #cell(total_amount)="data">
          {{ formatPrice(data.value) }}
        </template>
      </b-table>

      <!-- Custom table using Bootstrap Vue -->
      <!--
            <b-table-simple small borderless responsive>
              <b-thead>
                <b-tr>
                  <b-th>Order Name</b-th>
                  <b-th>Customer Company</b-th>
                  <b-th>Customer Name</b-th>
                  <b-th>Order Date</b-th>
                  <b-th>Delivered Amount</b-th>
                  <b-th>Total Amount</b-th>
                </b-tr>
              </b-thead>
              <b-tbody v-for="(item,index) in orders" :key="index">
                <b-tr>
                  <b-td class="seprateOrderProduct" ">{{ item.order_name }}</b-td>
                  <b-td rowspan="2">{{ item.customer_company }}</b-td>
                  <b-td rowspan="2">{{ item.customer_name }}</b-td>
                  <b-td rowspan="2">{{ item.order_date }}</b-td>
                  <b-td rowspan="2">{{ formatPrice(item.delivered_amount) }}</b-td>
                  <b-td rowspan="2">{{ formatPrice(item.total_amount) }}</b-td>
                </b-tr>
                <b-tr>
                  <b-td>{{ item.product_name }}</b-td>
                </b-tr>
              </b-tbody>
            </b-table-simple>
            -->

      <b-pagination
          v-model="page"
          :total-rows="totalElements"
          :per-page="pageSize"
          prev-text="Prev"
          next-text="Next"
          @change="handlePageChange">
      </b-pagination>
    </div>
  </div>
</template>

<script>
import FetchOrdersService from "@/services/FetchOrdersService";
import moment from 'moment';


export default {
  name: "Orders",
  data() {
    return {
      orders: [],
      fields: ["order_name", "customer_name", "customer_name", "order_date", "delivered_amount", "total_amount"],

      // custom leve naming for b-table fields/columns
      /*fields: [
        {
          key: 'order_name',
          label: 'Order Name'
        },
        {
          key: 'customer_company',
          label: 'Customer Company'
        }, {
          key: 'customer_name',
          label: 'Customer Name'
        }, {
          key: 'order_date',
          label: 'Order Date'
        }, {
          key: 'delivered_amount',
          label: 'Delivered Amount'
        }, {
          key: 'total_amount',
          label: 'Total Amount'
        }
      ],*/

      grand_total_amount: 0.0,
      searchPartOrderOrProductName: "",
      orderDateRange: "",
      page: 1,
      totalElements: 0,
      pageSize: 5,
    };
  },
  watch: {
    'orderDateRange': function () {
      this.page = 1
      this.getOrders()
    }
  },
  methods: {
    formatPrice(value) {
      if (typeof value !== "number") {
        return value;
      }
      const formatter = new Intl.NumberFormat('en-AU', {
        style: 'currency',
        currency: 'AUD',
        minimumFractionDigits: 0
      });
      return formatter.format(value);
    },
    formatOrderDate(value) {
      return moment(String(value)).format('MMM Do, YYYY hh:mm A');
    },
    getOrders() {
      const params = this.getRequestParams(this.page, this.pageSize, this.searchPartOrderOrProductName, this.orderDateRange);
      FetchOrdersService.getAll(params)
          .then((response) => {
            this.orders = response.data.data
            this.grand_total_amount = response.data.grand_total_amount
            this.page = response.data.page
            this.pageSize = response.data.pageSize
            this.totalElements = response.data.totalElements
          })
          .catch((e) => {
            console.log(e);
          });
    },

    handlePageChange(value) {
      this.page = value;
      this.getOrders();
    },

    formatRequestOrderDate(value) {
      return moment(String(value)).format('DD-MM-YYYY');
    },

    getRequestParams(page, pageSize, searchPartOrderOrProductName, orderDateRange) {
      let params = {};

      if (page) {
        params["page"] = page;
      }

      if (pageSize) {
        params["pageSize"] = pageSize;
      }

      if (searchPartOrderOrProductName) {
        params["orderNameOrProduct"] = searchPartOrderOrProductName;
      }
      if (orderDateRange) {
        params["startDate"] = this.formatRequestOrderDate(orderDateRange[0]);
        params["endDate"] = this.formatRequestOrderDate(orderDateRange[1]);
      }

      return params;
    },
  },
  mounted() {
    this.getOrders();
  },
}
</script>
<style scoped>
.seprateOrderProduct {
  white-space: pre-wrap !important;
}
</style>