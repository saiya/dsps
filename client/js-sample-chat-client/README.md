# DSPS JS client sample App

## How to run

1. `yarn --cwd ../js install`
2. `yarn --cwd ../js link`
3. `yarn --cwd ../js build`
4. `yarn link @dsps/client`
5. `yarn install`
6. Start [dsps server](../../server) in port 3000
    - e.g. `go run main.go` in the `../../server` directory
7. Start this app with `yarn start`
    - It opens [http://localhost:3001](http://localhost:3001) in your browser.
