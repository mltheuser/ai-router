plugins {
    kotlin("jvm") version "2.2.0"
    kotlin("plugin.serialization") version "2.2.0"
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

    // Serialization
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.7.3")

    // Test framework
    testImplementation("org.junit.jupiter:junit-jupiter:5.10.2")
    testImplementation("org.jetbrains.kotlin:kotlin-test")
}

kotlin {
    jvmToolchain(21)
}

tasks.test {
    useJUnitPlatform()
    testLogging.exceptionFormat = org.gradle.api.tasks.testing.logging.TestExceptionFormat.FULL
}

publishing {
    publications {
        create<MavenPublication>("maven") {
            from(components["java"])
        }
    }
}
