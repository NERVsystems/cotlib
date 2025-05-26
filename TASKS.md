- [x] Tighten wildcard validation to reject embedded '*'
- [x] Use context-based logging throughout
- [x] Add tests for embedded wildcard rejection
# Full Schema Coverage TODO
<!-- Remaining detail schemas have been embedded and compiled -->
- [x] Add validation hooks in Event.ValidateAt or during XML unmarshalling
- [x] Integrate remaining top-level schemas
  - Drawing shape schemas (Circle, Free Form, Rectangle, Telestration)
  - [x] Geo Fence and Range & Bearing schemas
  - [x] Route schemas (Route.xsd, tak-route.xsd)
  - [x] Marker schemas (2525, Icon Set, Spot)
- [x] Develop comprehensive tests for all new detail extensions
- [ ] Expand benchmarks for representative schemas

## Schema Integration Groups

Break down the remaining XSD work into manageable task groups:

### 1. Chat and Messaging Schemas
- [x] Embed `__chat.xsd`, `__chatreceipt.xsd`, and `chatgrp.xsd` in `validator/schemas/details/`.
- [x] Update `validator.go` to compile and register these schemas.
- [x] Add validation tests covering chat message details.

### 2. Geofence and Drawing Schemas
<!-- Completed -->

### 3. Group and User Schemas
<!-- Completed -->

### 4. Media and Attachment Schemas
- [x] Test validation of video, file share, and link attachments.

### 5. Environment and Mission Schemas
- [x] Add `environment.xsd`, `mission.xsd`, `precisionlocation.xsd`, and `takv.xsd`.
- [x] Implement tests focusing on mission planning and environment details.

### 6. Miscellaneous Schemas
- [x] Finalize `color.xsd` and `hierarchy.xsd` integration.
