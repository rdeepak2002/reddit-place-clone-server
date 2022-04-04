echo "Running release script..."
cd client && yarn add -g rollup && yarn install && yarn build
echo "Done running release script!"