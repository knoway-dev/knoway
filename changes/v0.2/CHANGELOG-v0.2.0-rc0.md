
# v0.2.0-rc0 Change logs

## Change since v0.1.2

### Changes by Kind

#### Bug

- feat: dedicated denied error for denied(!61) by @nekoayaka.zhang
- fix: vllm error handling(!65) by @nekoayaka.zhang
- > fix stream missing content-type(!67) by @kebe.liu
- Revert "feat: support metering usages for images generations"(!70) by @nekoayaka.zhang
- update ci image(!72) by @nicole.li
- feat: add rate limits(!74) by @nicole.li
- fix status equal(!76) by @nicole.li
- feat: add lb for model route(!77) by @xiaowu.zhu
- >fix base cluster not found & image chat not found(!79) by @nicole.li
- + support config listener via config file(!80) by @kebe.liu
- + add config_dump endpoint for debug(!81) by @kebe.liu
- feat: supported ratelimit redis(!82) by @nicole.li
- fix(fallback): not handling default LB policy & not handling invalid content-type with errored status code(!84) by @nekoayaka.zhang
- > fix duplicated requests send due to fallback(!87) by @kebe.liu


#### Feature

- feat: FromString util support ptr(!62) by @nekoayaka.zhang
- feat: image listener(!69) by @nekoayaka.zhang
- feat: size config(!71) by @nekoayaka.zhang
- feat: added ModelRoute CRD(!73) by @nekoayaka.zhang
- chore(route): align route.proto fields with new LB and retry ModelRoute CRD design(!75) by @nekoayaka.zhang
- feat(controller): model route(!78) by @nekoayaka.zhang
- feat: fallback(!83) by @nekoayaka.zhang



