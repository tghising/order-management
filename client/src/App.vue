<template>
  <div id="app">
    <nav class="navbar navbar-expand-lg navbar-light bg-light">
      <div class="container-fluid">
        <div class="collapse navbar-collapse" id="navbarSupportedContent">
          <ul class="navbar-nav me-auto mb-2 mb-lg-0">
            <li class="nav-item">
              <a class="nav-link" aria-current="page" href="/home">Home</a>
            </li>
            <li class="nav-item">
              <a class="nav-link" href="/about">About Us</a>
            </li>
            <li class="nav-item">
              <a class="nav-link" href="/customers">Customers</a>
            </li>
            <li class="nav-item">
              <a class="nav-link active" href="/orders">Orders</a>
            </li>
          </ul>
        </div>
      </div>
    </nav>

    <div class="container">
      <router-view/>
    </div>
  </div>

</template>
<script>
export default {
  name: "app"
};
</script>
