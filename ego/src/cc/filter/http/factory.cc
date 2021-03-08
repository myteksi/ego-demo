// Copyright 2020-2021 Grabtaxi Holdings PTE LTE (GRAB), All rights reserved.
//
// Use of this source code is governed by the Apache License 2.0 that can be
// found in the LICENSE file

#include <string>

#include "envoy/registry/registry.h"

#include "extensions/filters/http/common/factory_base.h"

#include "ego/src/cc/filter/http/filter.pb.validate.h"
#include "filter.h"

namespace Envoy {
namespace Server {
namespace Configuration {

// GoHttpFilterCf is the config factory registered with envoy core
//
class GoHttpFilterCf
    : public Envoy::Extensions::HttpFilters::Common::FactoryBase<ego::http::Settings,
                                                                 ego::http::SettingsPerRoute> {
public:
  GoHttpFilterCf() : FactoryBase(Http::GoHttpConstants::get().FilterName) {}

  Http::FilterFactoryCb
  createFilterFactoryFromProtoTyped(const ego::http::Settings& settings,
                                    const std::string& stats_prefix,
                                    Server::Configuration::FactoryContext& context) {

    auto cfg = std::make_shared<Http::GoHttpFilterConfig>(
        settings,
        context.scope().createScope(fmt::format(
            "{}{}.{}.", stats_prefix, Http::GoHttpConstants::get().FilterName, settings.filter())));

    Secret::GenericSecretConfigProviderSharedPtr secret_provider = nullptr;

    // Seems like it's possible to hit a pure virtual call when calling
    // Envoy::Secret::SecretManagerImpl::findOrCreateGenericSecretProvider during filter
    // construction with a file-based SDS secret: https://github.com/envoyproxy/envoy/issues/12013
    // So if we want to config static resource then setting it via static_resources, instead of
    //    sds_config:
    //      path: /etc/envoy/secret-resource.yaml

    // A filter can be configured without secret
    if (settings.has_sds_secret_config()) {
      // Follow this config logic
      // https://github.com/envoyproxy/envoy/blob/v1.14.1/source/extensions/transport_sockets/tls/context_config_impl.cc#L61
      // For working with static resource
      if (settings.sds_secret_config().has_sds_config()) {
        secret_provider =
            context.clusterManager()
                .clusterManagerFactory()
                .secretManager()
                .findOrCreateGenericSecretProvider(settings.sds_secret_config().sds_config(),
                                                   settings.sds_secret_config().name(),
                                                   context.getTransportSocketFactoryContext());
      } else {
        secret_provider = context.clusterManager()
                              .clusterManagerFactory()
                              .secretManager()
                              .findStaticGenericSecretProvider(settings.sds_secret_config().name());
      }
    }

    auto cgo_proxy = std::make_shared<Http::CgoProxyImpl>();

    return [cfg, &context, secret_provider,
            cgo_proxy](Http::FilterChainFactoryCallbacks& callbacks) -> void {
      auto span_group = std::make_unique<Envoy::Http::SpanGroup>();
      auto filter = new Http::GoHttpFilter(cfg, context.api(), secret_provider, cgo_proxy, std::move(span_group));
      callbacks.addStreamFilter(filter->ref());
    };
  }

  Router::RouteSpecificFilterConfigConstSharedPtr
  createRouteSpecificFilterConfigTyped(const ego::http::SettingsPerRoute& settings,
                                       Server::Configuration::ServerFactoryContext&,
                                       ProtobufMessage::ValidationVisitor&) {
    return std::make_shared<const Http::GoHttpRouteSpecificFilterConfig>(settings);
  }
};

/**
 * Static registration for the GoHttp filter. @see RegisterFactory
 */
REGISTER_FACTORY(GoHttpFilterCf, Server::Configuration::NamedHttpFilterConfigFactory);

} // namespace Configuration
} // namespace Server
} // namespace Envoy
