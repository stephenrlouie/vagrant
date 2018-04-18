# -*k mode: ruby -*-
# vi: set ft=ruby :

$num_clusters = 3
if ENV["NUM_CLUSTERS"] && ENV["NUM_CLUSTERS"].to_i > 0
    $num_clusters = ENV["NUM_CLUSTERS"].to_i
end

$box = ENV["VM_NAME"] || "intelligent-edge-admin/centos-k8s-1.10.0"
$box_version = ENV["VM_VERSION"] || "1.2.0"

$central_cluster_coords = (ENV["CENTRAL_CLUSTER_COORDS"] || "55.692770,12.598624").split(/\s*,\s*/)
$edge_cluster_coords = (ENV["EDGE_CLUSTER_COORDS"] || "55.680770,12.543006,55.664023,12.610126").split(/\s*,\s*/)

$svc_read_interval_secs = 2
$svc_push_interval_secs = 4

def provision_vm(config, vm_name, i)
    config.vm.hostname = vm_name
    config.vm.synced_folder ".", "/vagrant", disabled: true
    config.vm.box = $box
    config.vm.box_version = $box_version
    ip = "172.16.7.#{i+100}"
    config.vm.network :private_network, ip: ip
    config.vm.provision :shell, inline: "ifup eth1"
    config.vm.provision "shell", path: "scripts/reset-kube-config.sh", env: {"MYIP" => ip}, privileged: true
    config.vm.provision "file", source: "scripts/pv1.yaml", destination: "/home/vagrant/pv1.yaml"
    config.vm.provision "file", source: "scripts/tiller.yaml", destination: "/home/vagrant/tiller.yaml"
    config.vm.provision "shell", path: "scripts/deploy-helm.sh",  privileged: true
    config.vm.provision "file", source: "provision_files/id_rsa", destination: "/home/vagrant/id_rsa"
    config.vm.provision "file", source: "optikon-dns/kube-dns-svc.yaml", destination: "/home/vagrant/.coredns/kube-dns-svc.yaml"
    config.vm.provision :shell, inline: "kubectl -n kube-system replace -f /home/vagrant/.coredns/kube-dns-svc.yaml"
    config.vm.provision "file", source: "optikon-dns/kube-dns-depl.yaml", destination: "/home/vagrant/.coredns/kube-dns-depl.yaml"
    config.vm.provision :shell, inline: "kubectl -n kube-system replace -f /home/vagrant/.coredns/kube-dns-depl.yaml --force"
end

Vagrant.configure("2") do |config|

    if $edge_cluster_coords.length != (2 * ($num_clusters-1))
        raise Vagrant::Errors::VagrantError.new, "Incorrect number of edge cluster coordinates."
    end

    (1..$num_clusters).each do |i|
        if i == 1 # Do Central cluster first
            config.vm.define vm_name = "central", primary: true do |config|
                provision_vm(config, vm_name, i)
                config.vm.provision "file", source: "scripts/pv2.yaml", destination: "/home/vagrant/pv2.yaml"
                config.vm.provision "file", source: "provision_files/id_rsa.pub", destination: "/home/vagrant/id_rsa.pub"
                config.vm.provision "file", source: "optikon-dns/plugin/central/corefile.yaml", destination: "/home/vagrant/.coredns/corefile.yaml"
                config.vm.provision :shell, inline: "kubectl -n kube-system replace -f /home/vagrant/.coredns/corefile.yaml"
                config.vm.provision :shell, path: "scripts/trigger-coredns-reload.sh"
                if $num_clusters > 1
                    config.vm.provision "shell", path: "scripts/central-keys.sh", env: {"NUM_EDGE" => $num_clusters-1}, privileged: false
                    config.vm.provision "shell", path: "scripts/deploy-registry.sh", privileged: true
                end
            end
        else # Edge clusters
            config.vm.define vm_name = "%s-%01d" % ["edge", i-1] do |config|
                provision_vm(config, vm_name, i)
                config.vm.provision "shell", path: "scripts/send-kubeconfig.sh", :args => i-1,  privileged: true
                config.vm.provision "file", source: "optikon-dns/plugin/edge/corefile.yaml", destination: "/home/vagrant/.coredns/corefile.yaml"
                config.vm.provision :shell,
                    path: "scripts/replace-env-vars.sh",
                    env: {
                        "CENTRAL_IP" => "172.16.7.101",
                        "LON" => $edge_cluster_coords[2*(i-2)],
                        "LAT" => $edge_cluster_coords[2*(i-2)+1],
                        "SVC_READ_INTERVAL" => $svc_read_interval_secs,
                        "SVC_PUSH_INTERVAL" => $svc_push_interval_secs
                    }
                config.vm.provision :shell, inline: "kubectl -n kube-system replace -f /home/vagrant/.coredns/corefile.yaml"
                config.vm.provision :shell, path: "scripts/trigger-coredns-reload.sh"
                if i == 2 # Deploy echoserver on first edge cluster.
                    config.vm.provision "file", source: "manifests/echoserver.yaml", destination: "/home/vagrant/echoserver.yaml"
                    config.vm.provision :shell, inline: "kubectl create -f /home/vagrant/echoserver.yaml"
                end
            end
        end
    end

end
