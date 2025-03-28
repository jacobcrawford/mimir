{
  // The alert name is prefixed with the product name (eg. AlertName -> MimirAlertName).
  alertName(name)::
    $._config.alert_product + name,

  jobMatcher(job)::
    '%s=~"%s%s"' % [$._config.per_job_label, $._config.alert_job_prefix, formatJobForQuery(job)],

  jobNotMatcher(job)::
    '%s!~"%s%s"' % [$._config.per_job_label, $._config.alert_job_prefix, formatJobForQuery(job)],

  local formatJobForQuery(job) =
    if std.isArray(job) then '(%s)' % std.join('|', job)
    else if std.isString(job) then job
    else error 'expected job "%s" to be a string or an array, but it is type "%s"' % [job, std.type(job)],

  withRunbookURL(url_format, groups)::
    local update_rule(rule) =
      if std.objectHas(rule, 'alert')
      then rule {
        annotations+: {
          runbook_url: url_format % std.asciiLower(rule.alert),
        },
      }
      else rule;
    [
      group {
        rules: [
          update_rule(alert)
          for alert in group.rules
        ],
      }
      for group in groups
    ],

  withExtraLabelsAnnotations(groups)::
    local update_rule(rule) =
      if std.objectHas(rule, 'alert')
      then rule {
        annotations+: $._config.alert_extra_annotations,
        labels+: $._config.alert_extra_labels,
      }
      else rule;
    [
      group {
        rules: [
          update_rule(rule)
          for rule in group.rules
        ],
      }
      for group in groups
    ],

  alertRangeInterval(multiple)::
    ($._config.base_alerts_range_interval_minutes * multiple) + 'm',
}
