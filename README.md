# Reddit Place README Server

Server application to simulate r/place but for README.md files

## Requirements
Go (https://go.dev/doc/install)

## How to Use

1. Deploy a server application with this code (Heroku can directly deploy this Go project from a GitHub repository)

2. Add the following to your README.md file

```markdown
<img alt="image" src="https://website/static/image.png" width="300"/>
```

Note: Replace the "src" with the url your server is being hosted on

## Example

Add pixels here: https://reddit-place-readme-server.herokuapp.com/

Then refresh the page and notice the image below changing.

<img alt="image" src="https://reddit-place-readme-server.herokuapp.com/static/image.png" width="300"/> 

Note: The above example is using the following markdown
```markdown
<img alt="image" src="https://reddit-place-readme-server.herokuapp.com/static/image.png" width="300"/> 
```
