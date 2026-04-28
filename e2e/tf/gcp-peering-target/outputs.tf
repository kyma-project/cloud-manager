output "instance_ip_address" {
  value = google_compute_instance.vm_instance.network_interface[0].network_ip
}
output "vpc_id" {
  value = google_compute_network.vpc_network.name
}
output "project_id" {
  value = google_compute_network.vpc_network.project
}
