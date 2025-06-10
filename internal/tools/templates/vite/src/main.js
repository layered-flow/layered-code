import './style.css'

document.querySelector('#app').innerHTML = `
  <div>
    <h1>{{.AppName}}</h1>
    <p>Welcome to your new Vite app created with Layered Code.</p>
    <p>Edit <code>src/main.js</code> and save to test HMR</p>
  </div>
`