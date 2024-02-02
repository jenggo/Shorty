<!DOCTYPE html>
<html lang="en">

<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <meta http-equiv="X-UA-Compatible" content="ie=edge">
  <title>Login</title>
  <link href="/css/login.css" rel="stylesheet">
</head>

<body>
  <p id="notification"></p>
  <form id="loginForm" class="login" autocomplete="off">
    <input name="username" type="text" placeholder="Username">
    <input name="password" type="password" placeholder="Password">
    <button>Login</button>
  </form>
  <script src="https://cdn.jsdelivr.net/npm/bcryptjs@2/dist/bcrypt.min.js"></script>
  <script>
    const loginForm = document.getElementById("loginForm"), notificationElement = document.getElementById("notification"); var bcrypt = dcodeIO.bcrypt; loginForm.addEventListener("submit", async t => { t.preventDefault(); let e = new FormData(loginForm), n = { username: e.get("username"), password: await bcrypt.hash(e.get("password"), 10) }; try { let o = await fetch("/login", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(n) }); if (!o.ok) throw Error(`HTTP error! Status: ${o.status}`); let i = await o.json(); console.log(i), notificationElement.textContent = i.message, notificationElement.classList.add("success"), notificationElement.classList.remove("error") } catch (s) { console.error("Error during fetch:", s), notificationElement.textContent = "Invalid username or password.", notificationElement.classList.add("error"), notificationElement.classList.remove("success") } notificationElement.style.display = "block", setTimeout(() => { notificationElement.style.display = "none", location.reload() }, 5e3) });
  </script>
</body>

</html>
