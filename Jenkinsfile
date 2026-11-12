pipeline {
    agent any

    options {
        timestamps()
        disableConcurrentBuilds()
        skipDefaultCheckout(true)
        buildDiscarder(logRotator(numToKeepStr: '20'))
    }

    triggers {
        pollSCM('H/2 * * * *')
    }

    parameters {
        string(
            name: 'TEST_POSTGRES_IMAGE',
            defaultValue: 'postgres:18-alpine',
            description: 'Use the same PostgreSQL major version as the existing runtime database.'
        )
    }

    environment {
        APP_IMAGE = 'local/transaction-monitoring-system'
        CI_NETWORK = 'tms-ci'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm

                script {
                    env.GIT_COMMIT = sh(
                        script: 'git rev-parse HEAD',
                        returnStdout: true
                    ).trim()
                }
            }
        }

        stage('Integration tests') {
            steps {
                sh '''
                    set -eu

                    DB_CONTAINER="tms-ci-db-${BUILD_NUMBER}"

                    cleanup() {
                        docker rm -f "$DB_CONTAINER" >/dev/null 2>&1 || true
                    }
                    trap cleanup EXIT

                    docker network inspect "$CI_NETWORK" >/dev/null
                    docker rm -f "$DB_CONTAINER" >/dev/null 2>&1 || true

                    docker run -d \
                        --name "$DB_CONTAINER" \
                        --network "$CI_NETWORK" \
                        -e POSTGRES_USER=postgres \
                        -e POSTGRES_PASSWORD=ci_test_password \
                        -e POSTGRES_DB=main \
                        "$TEST_POSTGRES_IMAGE" >/dev/null

                    READY=0
                    for attempt in $(seq 1 60); do
                        if docker exec "$DB_CONTAINER" \
                            pg_isready -U postgres -d main >/dev/null 2>&1; then
                            READY=1
                            break
                        fi
                        sleep 2
                    done

                    if [ "$READY" -ne 1 ]; then
                        docker logs "$DB_CONTAINER"
                        exit 1
                    fi

                    mkdir -p .ci

                    cat > .ci/config.yaml <<EOF
jwt:
  secret: "ci-only-secret"
  expiry_in: "1800s"
postgres_db:
  host: "${DB_CONTAINER}"
  port: "5432"
  user: "postgres"
  password: "ci_test_password"
  db_name: "main"
http_server:
  address: "127.0.0.1:8082"
  timeout: "4s"
  idle_timeout: "60s"
tcp_server:
  address: "127.0.0.1:8084"
  idle_timeout: "600s"
EOF

                    CONFIG_PATH="$WORKSPACE/.ci/config.yaml" \
                    JWT_SECRET="ci-only-secret" \
                    POSTGRES_PASSWORD="ci_test_password" \
                        go test -count=1 ./...
                '''
            }
        }

        stage('Static analysis') {
            steps {
                sh 'go vet ./...'
            }
        }

        stage('Build Docker image') {
            steps {
                sh '''
                    set -eu
                    docker build --pull \
                        -t "$APP_IMAGE:$GIT_COMMIT" \
                        .
                '''
            }
        }

        stage('Deploy') {
            steps {
                withCredentials([
                    string(credentialsId: 'tms-jwt-secret', variable: 'JWT_SECRET'),
                    string(credentialsId: 'tms-postgres-password', variable: 'POSTGRES_PASSWORD')
                ]) {
                    sh '''
                        set -eu

                        IMAGE="$APP_IMAGE"
                        COMPOSE_FILE="deploy/compose.yaml"

                        PREVIOUS_ID="$(
                            docker inspect --format '{{.Image}}' tms-app 2>/dev/null || true
                        )"

                        deploy_image() {
                            IMAGE_REF="$1"
                            APP_IMAGE="$IMAGE_REF" \
                            JWT_SECRET="$JWT_SECRET" \
                            POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
                                docker compose -p tms -f "$COMPOSE_FILE" \
                                up -d --no-build --force-recreate
                        }

                        wait_healthy() {
                            CID="$(docker compose -p tms -f "$COMPOSE_FILE" ps -q app)"
                            [ -n "$CID" ] || return 1

                            for attempt in $(seq 1 45); do
                                STATUS="$(
                                    docker inspect \
                                      --format '{{.State.Health.Status}}' \
                                      "$CID" 2>/dev/null || echo missing
                                )"

                                [ "$STATUS" = "healthy" ] && return 0
                                [ "$STATUS" = "unhealthy" ] && return 1
                                sleep 2
                            done

                            return 1
                        }

                        rollback() {
                            if [ -n "$PREVIOUS_ID" ]; then
                                docker tag "$PREVIOUS_ID" "$IMAGE:rollback"

                                if ! deploy_image "$IMAGE:rollback"; then
                                    return 1
                                fi

                                if ! wait_healthy; then
                                    docker logs tms-app || true
                                    return 1
                                fi

                                docker tag "$PREVIOUS_ID" "$IMAGE:current"
                                echo "Previous application image restored."
                            else
                                docker compose -p tms -f "$COMPOSE_FILE" down
                                echo "First deployment failed; application container removed."
                            fi
                        }

                        if ! deploy_image "$IMAGE:$GIT_COMMIT"; then
                            echo "Deployment command failed; attempting rollback."
                            rollback || echo "Rollback attempt failed."
                            exit 1
                        fi

                        if ! wait_healthy; then
                            echo "New container did not become healthy."
                            docker logs tms-app || true
                            rollback || echo "Rollback attempt failed."
                            exit 1
                        fi

                        docker tag "$IMAGE:$GIT_COMMIT" "$IMAGE:current"
                        echo "Deployment completed and both TCP ports are listening."
                    '''
                }
            }
        }
    }

    post {
        always {
            sh '''
                docker rm -f "tms-ci-db-${BUILD_NUMBER}" >/dev/null 2>&1 || true
            '''
        }
        success {
            echo 'CI/CD completed successfully.'
        }
        failure {
            echo 'Pipeline failed. If deployment had started, check rollback logs.'
        }
    }
}
