- [x] Tighten wildcard validation to reject embedded '*'
- [x] Use context-based logging throughout
- [x] Add tests for embedded wildcard rejection
# Full Schema Coverage TODO
- [ ] Embed and compile remaining detail schemas
  - __chat.xsd
  - __chatreceipt.xsd
  - __geofence.xsd
  - __group.xsd
  - __serverdestination.xsd
  - __video.xsd
  - archive.xsd
  - attachment_list.xsd
  - chatgrp.xsd
  - color.xsd
  - environment.xsd
  - fileshare.xsd
  - fillColor.xsd
  - height.xsd
  - height_unit.xsd
  - hierarchy.xsd
  - labels_on.xsd
  - link.xsd
  - mission.xsd
  - precisionlocation.xsd
  - shape.xsd
  - strokeColor.xsd
  - strokeWeight.xsd
  - takv.xsd
  - uid.xsd
  - usericon.xsd
- [ ] Add validation hooks in Event.ValidateAt or during XML unmarshalling
- [ ] Integrate remaining top-level schemas
  - Drawing shape schemas (Circle, Free Form, Rectangle, Telestration)
  - Geo Fence and Range & Bearing schemas
  - Route schemas (Route.xsd, tak-route.xsd)
  - Marker schemas (2525, Icon Set, Spot)
- [ ] Develop comprehensive tests for all new detail extensions
- [ ] Expand benchmarks for representative schemas

## Schema Integration Groups

Break down the remaining XSD work into manageable task groups:

### 1. Chat and Messaging Schemas
- [x] Embed `__chat.xsd`, `__chatreceipt.xsd`, and `chatgrp.xsd` in `validator/schemas/details/`.
- [x] Update `validator.go` to compile and register these schemas.
- [x] Add validation tests covering chat message details.

### 2. Geofence and Drawing Schemas
- Integrate `__geofence.xsd`, `shape.xsd`, `strokeColor.xsd`, `strokeWeight.xsd`, `fillColor.xsd`, `height.xsd`, and `height_unit.xsd`.
- Extend validator tests for geofence and drawing elements.

### 3. Group and User Schemas
- Handle `__group.xsd`, `__serverdestination.xsd`, `uid.xsd`, `usericon.xsd`, and `labels_on.xsd`.
- Provide tests for group membership and user identity details.

### 4. Media and Attachment Schemas
- Embed `__video.xsd`, `attachment_list.xsd`, `fileshare.xsd`, `link.xsd`, and `archive.xsd`.
- Test validation of video, file share, and link attachments.

### 5. Environment and Mission Schemas
- Add `environment.xsd`, `mission.xsd`, `precisionlocation.xsd`, and `takv.xsd`.
- Implement tests focusing on mission planning and environment details.

### 6. Miscellaneous Schemas
- Finalize `color.xsd` and `hierarchy.xsd` integration.
- Review interactions with other detail schemas for completeness.
