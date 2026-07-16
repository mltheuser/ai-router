plugins {
    kotlin("jvm") version "2.2.0"
    kotlin("plugin.serialization") version "2.2.0"
    id("io.gitlab.arturbosch.detekt") version "1.23.8"
    `java-library`
    `maven-publish`
}

group = "ai.router"
version = "0.1.0"

repositories {
    mavenCentral()
}

val ktorVersion = "3.4.0"

dependencies {
    // HTTP client
    implementation("io.ktor:ktor-client-core:$ktorVersion")
    implementation("io.ktor:ktor-client-cio:$ktorVersion")
    implementation("io.ktor:ktor-client-content-negotiation:$ktorVersion")
    implementation("io.ktor:ktor-serialization-kotlinx-json:$ktorVersion")

    // Serialization. `api`: JsonObject and the @Serializable models sit in
    // the SDK's public signatures.
    api("org.jetbrains.kotlinx:kotlinx-serialization-json:1.7.3")

    // Test framework
    testImplementation("org.junit.jupiter:junit-jupiter:5.10.2")
    testImplementation("org.jetbrains.kotlin:kotlin-test")

    // Lint formatting rules (folds ktlint into detekt — no separate ktlint plugin)
    detektPlugins("io.gitlab.arturbosch.detekt:detekt-formatting:1.23.8")
}

kotlin {
    jvmToolchain(21)
    // Require an explicit visibility modifier and explicit return type on every
    // public/protected declaration — keeps the published API surface intentional.
    explicitApi = org.jetbrains.kotlin.gradle.dsl.ExplicitApiMode.Strict
    compilerOptions {
        allWarningsAsErrors = true
    }
}

detekt {
    buildUponDefaultConfig = true
    config.setFrom(files("detekt.yml"))
    // Scan both main and test sources.
    source.setFrom(files("src/main/kotlin", "src/test/kotlin"))
}

tasks.withType<io.gitlab.arturbosch.detekt.Detekt>().configureEach {
    // Fail the build on any detekt finding.
    ignoreFailures = false
}

tasks.test {
    useJUnitPlatform()
    testLogging.exceptionFormat = org.gradle.api.tasks.testing.logging.TestExceptionFormat.FULL
    outputs.upToDateWhen { false } // disable test result caching
}

publishing {
    publications {
        create<MavenPublication>("maven") {
            from(components["java"])
        }
    }
}
