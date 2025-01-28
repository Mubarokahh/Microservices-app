def branch = env.BRANCH_NAME 

if (branch == 'postgres') {
    load 'postgres/Jenkinsfile'
} else if (branch == 'redis') {
    load 'redis/Jenkinsfile'
} else if (branch == 'voting-service') {
    load 'voting-service/Jenkinsfile'
} else if (branch == 'worker-service') {
    load 'worker-service/Jenkinsfile'
} else if (branch == 'results-service') {
    load 'results-service/Jenkinsfile'
} else if (branch == 'main') {
    echo "Triggering builds for service branches..."
    build(job: 'voting-app/postgres', wait: false)    // the Correct job name format
    build(job: 'voting-app/redis', wait: false)
    build(job: 'voting-app/voting-service', wait: false)
    build(job: 'voting-app/worker-service', wait: false)
    build(job: 'voting-app/results-service', wait: false)
} else {
    error "No valid Jenkinsfile found for the branch: ${branch}"
}