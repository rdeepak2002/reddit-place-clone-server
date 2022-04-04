echo "Running release script..."
cd client && npm install --global rollup && yarn && yarn build
echo "Done running release script!"