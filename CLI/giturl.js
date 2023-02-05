const process = require('process');
const repoUrl = require('get-repository-url');

npmurl = process.argv[2].split('/');

// takes a callback
repoUrl(npmurl[4], function(err, url) {
  console.log(url);
});