package ai.router.sdk

/**
 * Marks the SDK's request builders, so a nested builder block cannot call
 * the methods of an enclosing one by accident.
 */
@DslMarker
public annotation class AiRouterDsl
