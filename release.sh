echo "Running release script..."
cd client && npm install --global rollup && npm install --global rollup-plugin-svelte && yarn && yarn build
echo "Done running release script!"