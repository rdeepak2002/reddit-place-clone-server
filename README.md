# Reddit Place Clone Server

Server application to simulate r/place

## Requirements
Go (https://go.dev/doc/install)

## Get Started

1. Clone the repository

```shell
git clone --recurse-submodules -j8 https://github.com/rdeepak2002/reddit-place-clone-server.git
```

2. Deploy a server application with this code (Heroku can directly deploy this Go project from a GitHub repository, but you need to add the submodules buildpack and Go buildpack)

```shell
heroku buildpacks:add https://github.com/dmathieu/heroku-buildpack-submodules -i 1
```

3. Create a .env file with the credentials for a Redis connection (you can get a free instance from here: https://redis.com/)

```dotenv
REDIS_ADDRESS="redis-xxx.com:#####"
REDIS_PASSWORD="really_long_password_string"
GOOGLE_AUTH_CLIENT_ID="xxxxx.apps.googleusercontent.com"
```

## Example Embed in README.md

Add pixels here: https://reddit-place-clone-server.herokuapp.com/

Then refresh the page and notice the image below changing (note that it is blurrier than the one present on the web application due to the lack of CSS styling in GitHub README's).

<img alt="image" src="https://reddit-place-clone-server.herokuapp.com/static/image.png" style="border: dotted black; width: 300px; height: 300px; image-rendering: pixelated; image-rendering: -moz-crisp-edges; image-rendering: crisp-edges;"/> 

The above example is using the following markdown:

```markdown
<img alt="image" src="https://reddit-place-clone-server.herokuapp.com/static/image.png" style="border: dotted black; width: 300px; height: 300px; image-rendering: pixelated; image-rendering: -moz-crisp-edges; image-rendering: crisp-edges;"/> 
```

## Update Web Application Code

```shell
git submodule update --remote 
```

## Recommended Pre-Commit Git Hooks

Create a file in .git/hooks with the following content:

```shell
#!/bin/sh
git submodule update --remote
git add .
```

Make the script executable with the following command:

```shell
sudo chmod 777 .git/hooks/pre-commit
```