# Migration Guide

This guide highlights the key changes introduced in recent releases and how to adapt code written for earlier versions of **cotlib**.

## Breaking Changes

- **Unexported security constants**: Limits such as `MaxDetailSize`, `MinStaleOffset` and `MaxStaleOffset` were unexported in v0.2.3. Use the setter functions (`SetMaxXMLSize`, `SetMaxElementDepth`, `SetMaxElementCount`, `SetMaxTokenLen` and `SetMaxValueLen`) to customise limits.
- **Detail structure updates**: Detail extensions now preserve XML attributes and unknown elements are captured in `Detail.Unknown`. Code that accessed plain strings should parse the new structures or handle the raw XML stored in these fields.
- **Time validation**: `stale` must be at least five seconds after `time` and no more than seven days later. Events outside this range fail validation.
- **Secure XML decoder**: DOCTYPE declarations are rejected and parser limits are enforced on every decode call. Ensure input XML conforms to these restrictions.

## New Security Defaults

Default parsing limits are configured in `init()`:

- Maximum XML size: **2 MB**
- Maximum element depth: **32**
- Maximum element count: **10,000**
- Maximum token length: **1,024 bytes**
- Maximum attribute or text length: **512 KiB**

Adjust these values with the setter functions if needed.

## Adapting Existing Code

1. Remove references to the old exported constants and rely on the setter functions to change limits.
2. When unmarshalling events, inspect `Detail.Unknown` for unrecognised extensions. Known extensions (such as `Chat` or `Remarks`) expose structured fields with preserved attributes.
3. Validate events with `Event.Validate()` to enforce the new time checks.
4. After processing, release events using `ReleaseEvent()` to return them to the pool.

Following these steps will ensure your application remains compatible with the latest library version.
