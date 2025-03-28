<h1 align="center">service template for <code>go</code></h1>
<div align="center">
  <a href="https://github.com/VU-ASE/service-template-go/releases/latest">Latest release</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="https://ase.vu.nl/docs/framework/glossary/service">About a service</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="https://ase.vu.nl/docs/framework/glossary/roverlib">About the roverlib</a>
  <br />
</div>
<br/>

**When building a service that runs on the Rover and should interface the ASE framework, you will most likely want to use a [roverlib](https://ase.vu.nl/docs/framework/glossary/roverlib). This is a Go template that incorporates [`roverlib-go`](https://github.com/VU-ASE/roverlib-go), meant to run on the Rover.**

## Initialize a Go service

Instead of cloning this repository, it is recommended to initialize this Go service using `roverctl` as follows:

```bash
roverctl service init go --name go-example --source github.com/author/go-example
```

Read more about using `roverctl` to initialize services [here](https://ase.vu.nl/docs/framework/Software/rover/roverctl/usage#initialize-a-service).


